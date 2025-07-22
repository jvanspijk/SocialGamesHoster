
using API.DataAccess;
using API.DataAccess.Repositories;
using API.Hubs;
using API.Services;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;
using System.Text;
using System.Text.Json.Serialization;

namespace API;

public class Program
{
    public static void Main(string[] args)
    {
        //TODO: add these as env variables:
        string hostUrl = "http://localhost:8080";
        string[] corsUrls = ["http://web:8081", "http://localhost:8081"];

        Console.WriteLine("Booting up app");
        var builder = WebApplication.CreateBuilder(args);
        var services = builder.Services;

        // Add services to the container.

        services.AddDbContext<APIDatabaseContext>(options =>
            options.UseNpgsql(builder.Configuration.GetConnectionString("DefaultConnection")));

        services.AddDistributedMemoryCache();

        services.AddSession(options =>
        {
            options.IdleTimeout = TimeSpan.FromSeconds(10);
            options.Cookie.HttpOnly = true;
            options.Cookie.IsEssential = true;
        });

        services.AddControllers().AddJsonOptions(o => 
            o.JsonSerializerOptions.ReferenceHandler = ReferenceHandler.IgnoreCycles
        );

        // Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
        services.AddEndpointsApiExplorer();
        services.AddSwaggerGen();

        services.AddCors(options =>
        {
            options.AddPolicy("CorsPolicy", builder =>
            {
                builder.WithOrigins(corsUrls)
                       .AllowAnyMethod()
                       .AllowAnyHeader()
                       .AllowCredentials();
            });
        });

        services.AddAuthentication(JwtBearerDefaults.AuthenticationScheme)
        .AddJwtBearer(options =>
        {
            options.MapInboundClaims = false;
            options.TokenValidationParameters = new TokenValidationParameters
            {
                ValidateIssuer = true,
                ValidateAudience = false, // This will make it easier for our environment, and security is not of a big concern here
                ValidateLifetime = true,
                ValidateIssuerSigningKey = true,
                ValidIssuer = hostUrl,
                IssuerSigningKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes("Social-games-hoster_JWT_Security_Key"))
            };
        });

        // We use signalR for websockets
        // To track if users are online/offline
        // And to send updates if the admin performs actions
        services.AddSignalR();

        services.AddScoped<RoleRepository>();
        services.AddScoped<AuthService>();
        services.AddScoped<PlayerRepository>();
        services.AddScoped<AbilityRepository>();
        services.AddScoped<AuthService>();
        services.AddSingleton<RoundRepository>();



        var app = builder.Build();

        // Apply database migrations on startup
        using (var scope = app.Services.CreateScope())
        {
            var serviceProvider = scope.ServiceProvider;
            try
            {
                var context = serviceProvider.GetRequiredService<APIDatabaseContext>();
                Console.WriteLine("Applying database migrations...");
                context.Database.Migrate();
                Console.WriteLine("Database migrations applied successfully.");
            }
            catch (Exception ex)
            {
                ILogger logger = serviceProvider.GetRequiredService<ILogger<Program>>();
                logger.LogError(ex, "An error occurred while applying database migrations.");
                throw;
            }
        }

        app.UseCors("CorsPolicy");

        app.UseDefaultFiles();

        // Configure the HTTP request pipeline.
        if (app.Environment.IsDevelopment())
        {
            app.UseSwagger();
            app.UseSwaggerUI();
        }

        app.UseExceptionHandler("/Error");

        app.UseRouting();

        app.UseSession();

        app.UseAuthorization();

        app.MapControllers();

        app.MapHub<PresenceHub>("/presenceHub");

        app.Run();
        Console.WriteLine("Done booting up");
    }
}

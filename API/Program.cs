using API.DataAccess;
using API.DataAccess.Repositories;
using API.Features.Abilities;
using API.Features.Authentication;
using API.Features.Players;
using API.Features.Roles;
using API.Hubs;
using API.Logging;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.HttpLogging;
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

        // Configure logging
        builder.Logging.ClearProviders();
        builder.Logging.AddConsole();
        builder.Logging.AddFileLogger("logs/api.log", 10);

        services.AddHttpLogging(options =>
        {
            options.LoggingFields = HttpLoggingFields.RequestMethod
                                  | HttpLoggingFields.RequestPath
                                  | HttpLoggingFields.RequestQuery
                                  | HttpLoggingFields.RequestBody
                                  | HttpLoggingFields.ResponseStatusCode
                                  | HttpLoggingFields.ResponseBody;

            options.RequestBodyLogLimit = 4096;
            options.ResponseBodyLogLimit = 4096;
            options.RequestHeaders.Clear();
            options.ResponseHeaders.Clear();
            options.CombineLogs = true;
        });

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

        services.AddScoped<RoleRepository>()
            .AddScoped<PlayerRepository>()
            .AddScoped<AbilityRepository>()
            .AddSingleton<RoundRepository>();

        services.AddScoped<AbilityService>()
            .AddScoped<AuthService>()
            .AddScoped<PlayerService>()
            .AddScoped<RoleService>();

        var app = builder.Build();

        // Apply database migrations on startup
        using (var scope = app.Services.CreateScope())
        {
            var serviceProvider = scope.ServiceProvider;
            try
            {
                APIDatabaseContext context = serviceProvider.GetRequiredService<APIDatabaseContext>() 
                    ?? throw new InvalidOperationException("Couldn't get database context for applying migrations.");
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
            app.UseHttpLogging();
            app.UseDeveloperExceptionPage();
        }
        else
        {
            app.UseExceptionHandler("/Error");
        }

        app.UseRouting();

        app.UseSession();

        app.UseAuthorization();

        app.MapControllers();

        app.MapHub<PresenceHub>("/presenceHub");

        app.Run();
        Console.WriteLine("Done booting up");
    }
}

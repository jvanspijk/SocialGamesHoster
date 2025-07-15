
using API.DataAccess;
using API.DataAccess.Repositories;
using API.Hubs;
using API.Models;
using API.Services;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.Identity;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;
using System.Text;

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

        //services.AddIdentity<User, IdentityRole>(options =>
        //{
        //    options.Password.RequiredLength = 0;
        //    options.Password.RequireUppercase = false;
        //    options.Password.RequireLowercase = false;
        //    options.Password.RequireDigit = false;
        //})
        //.AddEntityFrameworkStores<APIDatabaseContext>();

        services.AddControllers();
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

        services.AddSingleton<RoleRepository>();
        services.AddSingleton<AuthService>();
        services.AddSingleton<UserRepository>();

        var app = builder.Build();
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

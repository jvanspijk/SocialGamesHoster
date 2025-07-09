
using API.Hubs;
using API.Models;
using API.Services;
using Microsoft.Extensions.DependencyInjection;

namespace API;

public class Program
{
    public static void Main(string[] args)
    {
        string[] corsUrls = ["http://web:51144", "http://localhost:51144", "http://172.18.0.3:51144"];

        Console.WriteLine("Booting up app");
        var builder = WebApplication.CreateBuilder(args);
        var services = builder.Services;

        // Add services to the container.

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


        // We use signalR for websockets
        // To track if users are online/offline
        // And to send updates if the admin performs actions
        //test
        services.AddSignalR();

        services.AddSingleton<UserService>();

        var app = builder.Build();
        app.UseCors("CorsPolicy");

        app.UseDefaultFiles();

        // Configure the HTTP request pipeline.
        if (app.Environment.IsDevelopment())
        {
            app.UseSwagger();
            app.UseSwaggerUI();
        }

        app.UseRouting();

        app.UseAuthorization();

        app.MapControllers();

        app.MapHub<PresenceHub>("/presenceHub");

        app.Run();
        Console.WriteLine("Done booting up");
    }
}

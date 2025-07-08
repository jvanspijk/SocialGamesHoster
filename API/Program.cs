
using API.Hubs;
using API.Models;
using API.Services;

namespace API;

public class Program
{
    public static void Main(string[] args)
    {
        Console.WriteLine("Booting up app");
        var builder = WebApplication.CreateBuilder(args);

        // Add services to the container.

        builder.Services.AddControllers();
        // Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
        builder.Services.AddEndpointsApiExplorer();
        builder.Services.AddSwaggerGen();

        // We use signalR for websockets
        // To track if users are online/offline
        // And to send updates if the admin performs actions
        //test
        builder.Services.AddSignalR();

        builder.Services.AddSingleton<UserService>();

        var app = builder.Build();

        app.UseDefaultFiles();

        // Configure the HTTP request pipeline.
        if (app.Environment.IsDevelopment())
        {
            app.UseSwagger();
            app.UseSwaggerUI();
        }

        app.UseAuthorization();

        app.MapControllers();

        app.MapHub<PresenceHub>("/presenceHub");

        app.Run();
        Console.WriteLine("Done booting up");
    }
}

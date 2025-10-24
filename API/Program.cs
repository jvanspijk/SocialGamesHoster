using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Features.Authentication;
using API.Features.Rounds.Hubs;
using API.Logging;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.HttpLogging;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.Tokens;
using Scalar.AspNetCore;
using System.Text;
using System.Text.Json.Serialization;

namespace API;
// Using scalar: http://localhost:9090/scalar
// OpenAPI scheme hosted at: http://localhost:9090/openapi/v1.json
// TODO:
// - Assign random roles to players
//   - How do we come up with the general logic for how many roles of each type are assigned?
//   Maybe with a percentage of particpants that can have a role per role in the ruleset? But then some roles are mandatory, or can always only have 1.

// - Fix login for players
//      - Use local IP address to identify players
// - Admin: force logout users, (decouple IP from user)
// - Fix login for admins
// - Change participants in active game sessions
//   - Remove players from game sessions
// - Always inject repository interfaces instead of concrete repositories
// - Testing project with unit and/or integration tests
// - Performance testing

// Bugs/issues:
// - Minor issue: cancelling a round increments the round number. This might be an issue for games where there's a fixed number of rounds.
// - Rounds can only be created with a timer that is started at creation and overrides the previous one. There's no way to create a round without a timer.
// - Admin name and password should be stored in the database, yet easily changeable. Maybe put it in the docker-compose environment variables.
public class Program
{
    public static void Main(string[] args)
    {
        //TODO: add these as env variables:
        string hostUrl = "http://localhost:9090";
        string[] corsUrls = ["http://web:9091", "http://localhost:9091"];

        Console.WriteLine("Booting up app");
        var builder = WebApplication.CreateBuilder(args);
        var services = builder.Services;
        var environment = builder.Environment;

        // Configure logging
        builder.Logging.ClearProviders();
        builder.Logging.AddConsole();
        builder.Logging.AddFileLogger("logs/requests.log", 10, LogLevel.Information, LogLevel.Information);
        builder.Logging.AddFileLogger("logs/errors.log", 5, LogLevel.Warning, LogLevel.Critical);

        services.AddHttpLogging(options =>
        {
            options.LoggingFields = HttpLoggingFields.RequestMethod
                                  | HttpLoggingFields.RequestPath
                                  | HttpLoggingFields.RequestQuery
                                  | HttpLoggingFields.RequestBody
                                  | HttpLoggingFields.ResponseStatusCode
                                  | HttpLoggingFields.ResponseBody;

            options.RequestBodyLogLimit = 512;
            options.ResponseBodyLogLimit = 512;
            options.RequestHeaders.Clear();
            options.ResponseHeaders.Clear();
            options.CombineLogs = true;
        });

        // Add services to the container.

        services.AddDbContext<APIDatabaseContext>(options =>
        {
            if (environment.IsDevelopment())
            {
                options.UseInMemoryDatabase("DevInMemoryDb");
            }
            else
            {
                options.UseNpgsql(builder.Configuration.GetConnectionString("DefaultConnection"));
            }
        });

        services.AddDistributedMemoryCache();

        services.AddSession(options =>
        {
            options.IdleTimeout = TimeSpan.FromSeconds(10);
            options.Cookie.HttpOnly = true;
            options.Cookie.IsEssential = true;
        });

        services.ConfigureHttpJsonOptions(options =>
        {
            options.SerializerOptions.ReferenceHandler = ReferenceHandler.IgnoreCycles;
        });

        // Learn more about configuring Swagger/OpenAPI at https://aka.ms/aspnetcore/swashbuckle
        services.AddEndpointsApiExplorer();
        services.AddOpenApi(options => options.AddSchemaTransformer(new NestedClassSchemaTransformer()));        

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

        services
            .AddHostedService<TimerEventNotifier>();

        // Dependency injection for concrete repositories
        // This should be replaced by IRepository<T> injections later
        services
            .AddScoped<AbilityRepository>()
            .AddScoped<PlayerRepository>()
            .AddScoped<RoleRepository>()
            .AddScoped<RoundRepository>()
            .AddScoped<RulesetRepository>()
            .AddScoped<GameSessionRepository>()
            .AddScoped<AuthService>();

        services
            .AddScoped<IRepository<Ability>, AbilityRepository>()
            .AddScoped<IRepository<Player>, PlayerRepository>()
            .AddScoped<IRepository<Role>, RoleRepository>()
            .AddScoped<IRepository<Round>, RoundRepository>()
            .AddScoped<IRepository<Ruleset>, RulesetRepository>()
            .AddScoped<IRepository<GameSession>, GameSessionRepository>();

        services
            .AddSingleton<RoundTimer>();

        var app = builder.Build();

        ApplyDatabaseMigrations(environment, app);

        app.UseCors("CorsPolicy");

        app.UseDefaultFiles();

        // Configure the HTTP request pipeline.
        if (app.Environment.IsDevelopment())
        {
            app.MapOpenApi();
            app.MapScalarApiReference(o => o
                .WithTheme(ScalarTheme.Alternate)
                .Layout = ScalarLayout.Classic
            );
            app.UseHttpLogging();
            app.UseDeveloperExceptionPage();
        }
        else
        {
            app.UseExceptionHandler("/Error");
        }

        app.UseRouting();

        app.UseSession();

        app.UseAuthentication();
        app.UseAuthorization();

        var apiGroup = app.MapGroup("/api");
        apiGroup.MapEndpoints();        
        apiGroup.MapHub<TimerHub>("/timerHub");
        apiGroup.MapGet("/health", () => Results.Ok("API is running")).WithTags("Health");

        // apiGroup.MapHub<PresenceHub>("/presenceHub");

        app.Run();
        Console.WriteLine("Done booting up");
    }

    private static void ApplyDatabaseMigrations(IWebHostEnvironment environment, WebApplication app)
    {
        using var scope = app.Services.CreateScope();
        var serviceProvider = scope.ServiceProvider;
        try
        {
            APIDatabaseContext context = serviceProvider.GetRequiredService<APIDatabaseContext>()
                ?? throw new InvalidOperationException("Couldn't get database context for applying migrations.");
            if (environment.IsDevelopment())
            {
                Console.WriteLine("Creating in-memory database...");
                context.Database.EnsureCreated();
                Console.WriteLine("In-memory database created successfully.");
            }
            else
            {
                Console.WriteLine("Applying database migrations...");
                context.Database.Migrate();
                Console.WriteLine("Database migrations applied successfully.");
            }
        }
        catch (Exception ex)
        {
            ILogger logger = serviceProvider.GetRequiredService<ILogger<Program>>();
            logger.LogError(ex, "An error occurred while applying database migrations.");
            throw;
        }
    }
}

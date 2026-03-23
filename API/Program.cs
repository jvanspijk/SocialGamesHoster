using API.DataAccess;

using API.Domain;
using API.Domain.Entities;
using API.Domain.Models;
using API.Features.Auth;
using API.Features.Timers;
using API.Logging;
using Microsoft.AspNetCore.Authentication.JwtBearer;
using Microsoft.AspNetCore.HttpLogging;
using Microsoft.AspNetCore.HttpOverrides;
using Microsoft.EntityFrameworkCore;
using Microsoft.IdentityModel.JsonWebTokens;
using Microsoft.IdentityModel.Tokens;
using Scalar.AspNetCore;
using System;
using System.Linq;
using System.Text;
using System.Text.Json.Serialization;

namespace API;
// Using scalar: http://localhost:9090/scalar
// OpenAPI scheme hosted at: http://localhost:9090/openapi/v1.json
// TODO:
// Big refactors:
// - Remove docker entirely. Use install and run script instead.
// v1:
// - Maybe add Open telemetry for performance logging
// - Fix login for admins
//      - store admin credentials in database or environment variables
// - Admin: force logout users, (decouple IP from user)
// - Change participants in active game sessions
//   - Remove players from game sessions
// - relationship class, relationship types are stored in ruleset e,g, neighbor, teammate, nemesis, etc.
// - Fix login for players
//      - login using player id (?)
//      - Use local IP address to identify players
// - Fix adjust timer endpoint:
//      - it assumes delta is negative
//      - Calculation of total time is off
// - GetTimerState should have a result for the case where there is no timer.
// - Solve todos in player repository
// - Search for more todos and fix them
// Wrong parameters gave a 200 OK response. Maybe svelte issue.

// v2:
// - Always inject repository interfaces instead of concrete repositories
// - Testing project with unit and/or integration tests
// - Performance testing
// - chats
// - roles have win conditions
// - IAbility interface so that a DM can create custom abilities for roles (will be god object)
//    - Parse abilities from JSON or scripting language instead so that users don't have to write code


// Optional:
// - Assign random roles to players
//   - How do we come up with the general logic for how many roles of each type are assigned?
//   Maybe with a percentage of particpants that can have a role per role in the ruleset? But then some roles are mandatory, or can always only have 1.

// Performance optimizations:
// - Streaming database results using IAsyncEnumerable where possible, especially for endpoints that return lists of data. <-- while the asp.net api result can return a stream, it needs to be handled client side too. There's also some latency involved in starting to return results, so it might not be worth it for smaller lists of data. But for larger lists, it can improve performance and reduce memory usage.
// - Use JSON source generator for serialization where possible (see GetAbility endpoint for example). <-- Does it save enough time?



// Bugs/issues:
// - Minor issue: cancelling a round increments the round number (not talking about the id). This might be an issue for games where there's a fixed number of rounds.
// - Rounds can only be created with a timer that is started at creation and overrides the previous one. There's no way to create a round without a timer.
// - Admin name and password should be stored in the database, yet easily changeable. Maybe put it in the docker-compose environment variables.
public class Program
{
    public static void Main(string[] args)
    {
        Console.WriteLine("Booting up app");
        var builder = WebApplication.CreateBuilder(args);

        if (string.IsNullOrEmpty(Environment.GetEnvironmentVariable("ASPNETCORE_ENVIRONMENT"))
            && string.IsNullOrEmpty(Environment.GetEnvironmentVariable("DOTNET_ENVIRONMENT")))
        {
            builder.Environment.EnvironmentName = Environments.Development;
        }

        builder.Configuration
            .AddJsonFile("Properties/appsettings.json", optional: true, reloadOnChange: true)
            .AddJsonFile($"Properties/appsettings.{builder.Environment.EnvironmentName}.json", optional: true, reloadOnChange: true);

        string hostUrl = builder.Configuration["HOST_URL"] ?? "http://localhost:9090";
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
                                  | HttpLoggingFields.ResponseStatusCode;

            options.RequestBodyLogLimit = 512;
            options.RequestHeaders.Clear();
            options.ResponseHeaders.Clear();
            options.CombineLogs = true;
        });

        // Add services to the container.

        bool isGeneratingDocs = IsDocumentGenerationRun(args);
        string? connectionString = builder.Configuration.GetConnectionString("DefaultConnection");

        services.AddDbContext<APIDatabaseContext>(options =>
        {
            if (string.IsNullOrWhiteSpace(connectionString))
            {
                throw new InvalidOperationException("Connection string 'DefaultConnection' not found.");
            }

            options.UseNpgsql(connectionString);
        });

        services.AddDistributedMemoryCache(); //For session state
        services.AddMemoryCache(); //For general caching

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

        services.AddEndpointsApiExplorer();
        services.AddOpenApi(options =>
        {
            options.CreateSchemaReferenceId = typeInfo => typeInfo.Type.FullName?.Replace("+", "_");
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

        // Dependency injection for concrete repositories
        // This should be replaced by IRepository<T> injections later
        services
            .AddScoped<AuthService>();

        services
            .AddScoped(typeof(IRepository<>), typeof(Repository<>));

        services
            .AddSingleton<RoundTimer>()
            .AddSingleton<TimerNotifier>();

        services.AddHostedService(provider => provider.GetRequiredService<TimerNotifier>());

        builder.WebHost.ConfigureKestrel(options =>
        {
            options.ListenAnyIP(9090);
        });

        var app = builder.Build();

        if (!isGeneratingDocs && !EF.IsDesignTime)
        {
            EnsureDatabaseConnection(app);
            ApplyDatabaseMigrations(app);
        }

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

        app.UseForwardedHeaders(new ForwardedHeadersOptions
        {
            ForwardedHeaders = ForwardedHeaders.XForwardedFor | ForwardedHeaders.XForwardedProto
        });

        app.UseRouting();

        app.UseSession();

        app.UseAuthentication();
        app.UseAuthorization();

        var apiGroup = app.MapGroup("/api");
        apiGroup.MapEndpoints();        
        apiGroup.MapGet("/health", () => Results.Ok("API is running")).WithTags("Health");
#if DEBUG
        apiGroup.AddEndpointFilter<DebugRequireSaveChangesFilter>();
#endif
        apiGroup.AddEndpointFilter<CacheInvalidatorFilter>();

        app.Run();
        Console.WriteLine("Done booting up");
    }

    private static void ApplyDatabaseMigrations(WebApplication app)
    {
        using var scope = app.Services.CreateScope();
        APIDatabaseContext context = scope.ServiceProvider.GetRequiredService<APIDatabaseContext>()
            ?? throw new InvalidOperationException("Couldn't get database context for applying migrations.");

        Console.WriteLine("Applying migrations...");
        context.Database.Migrate();
        Console.WriteLine("Database is up to date.");
    }

    private static void EnsureDatabaseConnection(WebApplication app)
    {
        using var scope = app.Services.CreateScope();
        APIDatabaseContext context = scope.ServiceProvider.GetRequiredService<APIDatabaseContext>()
            ?? throw new InvalidOperationException("Couldn't get database context for connectivity check.");

        if (!context.Database.CanConnect())
        {
            throw new InvalidOperationException("PostgreSQL is unavailable. Startup aborted.");
        }
    }

    private static bool IsDocumentGenerationRun(string[] args)
    {
        if (args.Contains("--get-document"))
        {
            return true;
        }

        if (Environment.GetCommandLineArgs().Any(a => a.Contains("dotnet-getdocument", StringComparison.OrdinalIgnoreCase)))
        {
            return true;
        }

        if (Environment.CommandLine.Contains("dotnet-getdocument", StringComparison.OrdinalIgnoreCase)
            || Environment.CommandLine.Contains("Microsoft.Extensions.ApiDescription.Tool", StringComparison.OrdinalIgnoreCase))
        {
            return true;
        }

        return AppDomain.CurrentDomain.FriendlyName.Contains("dotnet-getdocument", StringComparison.OrdinalIgnoreCase);
    }
}

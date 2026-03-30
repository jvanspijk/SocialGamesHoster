using API.DataAccess;

using API.Domain;
using API.Domain.Entities;
using API.Domain.Models;
using API.Features.Auth;
using API.Features.Timers;
using API.Filters;
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
// - Uniform error handling. In svelte, add a CreateFail function to ApiError so that we can return after !res.ok something like: return res.error.fail()
// v1:
// - Maybe add Open telemetry for performance logging
// - Fix login for admins
//      - store admin credentials in database or environment variables
// - Admin: force logout users, (decouple IP from user)
// - Change participants in active game sessions
//   - Remove players from game sessions
// - Fix login for players
//      - login using player id (?)
//      - Use local IP address to identify players
// - Fix adjust timer endpoint:
//      - it assumes delta is negative
//      - Calculation of total time is off
// - GetTimerState should have a result for the case where there is no timer.
// - Solve todos in player repository
// - Search for more todos and fix them
// - General chat
// - DM the DM
// - Hide button for the role info
// - Logout endpoitn and button
// - Management panels for deleting and updating resources
// Wrong parameters gave a 200 OK response. Maybe svelte issue.

// v2:
// - relationship class, relationship types are stored in ruleset e,g, neighbor, teammate, nemesis, etc.
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

#if DEBUG
        builder.Logging.AddSqliteLogger("Data Source=logs/debug_logs.db;");        
        EnsureLogDatabaseCreated();
#endif

        // Configure CORS
        services.AddCors(options =>
        {
            options.AddPolicy("Cors", policy =>
            {
                policy.SetIsOriginAllowed(_ => true)
                      .AllowAnyHeader()
                      .AllowAnyMethod()
                      .AllowCredentials();
            });
        });


        // Add services to the container.

        bool isGeneratingDocs = IsDocumentGenerationRun(args);
        string? connectionString = builder.Configuration.GetConnectionString("DefaultConnection");

#if DEBUG
        builder.Services.AddHttpContextAccessor();
        builder.Services.AddScoped<SqliteDbInterceptor>();

        builder.Services.AddDbContext<APIDatabaseContext>((sp, options) =>
        {
            options.UseNpgsql(builder.Configuration.GetConnectionString("DefaultConnection"))
                   .AddInterceptors(sp.GetRequiredService<SqliteDbInterceptor>());
        });
#else
 services.AddDbContext<APIDatabaseContext>(options =>
        {
            if (string.IsNullOrWhiteSpace(connectionString))
            {
                throw new InvalidOperationException("Connection string 'DefaultConnection' not found.");
            }

            options.UseNpgsql(connectionString);
        });
#endif

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

            options.Events = new JwtBearerEvents
            {
                OnMessageReceived = context =>
                {
                    if (context.Request.Cookies.TryGetValue("session_token", out var token))
                    {
                        context.Token = token;
                    }
                    return Task.CompletedTask;
                }
            };
        });

        services.AddAuthorizationBuilder()
            .AddPolicy("AdminOnly", policy => policy.RequireClaim("role", "admin"))
            .AddPolicy("PlayerOnly", policy => policy.RequireClaim("role", "player"));

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
            .AddSingleton<IGameTimer, GameTimer>()
            .AddSingleton<TimerNotifier>();

        services.AddHostedService(provider => provider.GetRequiredService<TimerNotifier>());

        builder.WebHost.ConfigureKestrel(options =>
        {
            options.ListenAnyIP(9090);
        });

        var app = builder.Build();
        app.UseCors("Cors");
        app.UseRouting();

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

        app.UseSession();

        app.UseAuthentication();
        app.UseAuthorization();

        var apiGroup = app.MapGroup("/api");        
        apiGroup.MapEndpoints();        
        apiGroup.MapGet("/health", () => Results.Ok("API is running")).WithTags("Health");
        apiGroup
            .AddEndpointFilter<HttpErrorFormattingFilter>()
            .AddEndpointFilter<CacheInvalidatorFilter>();
#if DEBUG
        // Ensure that we can read and inspect the http request
        apiGroup.AddEndpointFilter(async (context, next) => 
        {
            context.HttpContext.Request.EnableBuffering();
            return await next(context);
        });
        apiGroup
            .AddEndpointFilter<RequestLoggingFilter>()
            .AddEndpointFilter<ExceptionLoggingFilter>()
            .AddEndpointFilter<DebugRequireSaveChangesFilter>();
#endif

        app.Run();
        Console.WriteLine("Done booting up");
    }

    private static void EnsureLogDatabaseCreated()
    {
        SQLitePCL.Batteries.Init();
        const string dbPath = "logs/debug_logs.db";
        const string connectionString = $"Data Source={dbPath};";

        var folder = Path.GetDirectoryName(dbPath);
        if (!string.IsNullOrEmpty(folder) && !Directory.Exists(folder))
        {
            Directory.CreateDirectory(folder);
        }

        using var connection = new Microsoft.Data.Sqlite.SqliteConnection(connectionString);
        connection.Open();

        using var command = connection.CreateCommand();
        command.CommandText = @"
            PRAGMA journal_mode=WAL;
            PRAGMA synchronous=OFF;

            CREATE TABLE IF NOT EXISTS Requests (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                Endpoint TEXT,
                Method TEXT,
                RequestBody TEXT,
                StatusCode INTEGER,
                ElapsedMS REAL,
                TraceId TEXT
            );

            CREATE TABLE IF NOT EXISTS QueryLogs (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                TraceId TEXT, 
                QueryText TEXT,
                ElapsedMS REAL,
                Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
            );

            CREATE TABLE IF NOT EXISTS ErrorLogs (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
                TraceId TEXT,
                ErrorMethod TEXT,
                ExceptionType TEXT,
                Message TEXT,
                StackTrace TEXT,
                StackTraceHash TEXT,
                ExceptionSource TEXT,
                TargetSite TEXT,
                Endpoint TEXT
            );

            CREATE TABLE IF NOT EXISTS StackTraces (
                ID INTEGER PRIMARY KEY AUTOINCREMENT,
                StackTraceHash TEXT NOT NULL UNIQUE,
                StackTraceText TEXT NOT NULL,
                FirstSeenUtc DATETIME DEFAULT CURRENT_TIMESTAMP,
                LastSeenUtc DATETIME DEFAULT CURRENT_TIMESTAMP,
                SeenCount INTEGER NOT NULL DEFAULT 1
            );

            CREATE INDEX IF NOT EXISTS idx_requests_trace ON Requests (TraceId);
            CREATE INDEX IF NOT EXISTS idx_queries_request ON QueryLogs (TraceId);
            CREATE INDEX IF NOT EXISTS idx_errors_trace ON ErrorLogs (TraceId);
            CREATE INDEX IF NOT EXISTS idx_errors_timestamp ON ErrorLogs (Timestamp);
            CREATE INDEX IF NOT EXISTS idx_errors_stackhash ON ErrorLogs (StackTraceHash);
            CREATE INDEX IF NOT EXISTS idx_queries_timestamp ON QueryLogs (Timestamp);
        ";

        command.ExecuteNonQuery();

        EnsureColumnExists(connection, "ErrorLogs", "StackTraceHash", "TEXT");
        EnsureColumnExists(connection, "ErrorLogs", "ExceptionSource", "TEXT");
        EnsureColumnExists(connection, "ErrorLogs", "TargetSite", "TEXT");
        EnsureColumnExists(connection, "ErrorLogs", "ErrorMethod", "TEXT");

        Console.WriteLine($"Database created at: {Path.GetFullPath(connection.DataSource)}");
    }

    private static void EnsureColumnExists(Microsoft.Data.Sqlite.SqliteConnection connection, string tableName, string columnName, string columnType)
    {
        using var checkCommand = connection.CreateCommand();
        checkCommand.CommandText = $"PRAGMA table_info({tableName});";

        using var reader = checkCommand.ExecuteReader();
        while (reader.Read())
        {
            var existingColumnName = reader.GetString(1);
            if (string.Equals(existingColumnName, columnName, StringComparison.OrdinalIgnoreCase))
            {
                return;
            }
        }

        using var alterCommand = connection.CreateCommand();
        alterCommand.CommandText = $"ALTER TABLE {tableName} ADD COLUMN {columnName} {columnType};";
        alterCommand.ExecuteNonQuery();
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

        if (Environment.GetEnvironmentVariable("ASPNETCORE_HOSTINGSTARTUPASSEMBLIES")?.Contains("Microsoft.AspNetCore.OpenApi") == true)
            return true;

        return AppDomain.CurrentDomain.FriendlyName.Contains("dotnet-getdocument", StringComparison.OrdinalIgnoreCase);
    }
}

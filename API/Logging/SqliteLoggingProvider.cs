namespace API.Logging;

public class SqliteLoggingProvider(string connectionString) : ILoggerProvider
{
    public ILogger CreateLogger(string categoryName) => new SqliteLogger(connectionString);
    public void Dispose() { }
}

public static class SqliteLoggerExtensions
{
    public static ILoggingBuilder AddSqliteLogger(this ILoggingBuilder builder, string connectionString)
    {
        builder.Services.AddSingleton<ILoggerProvider>(new SqliteLoggingProvider(connectionString));
        return builder;
    }
}
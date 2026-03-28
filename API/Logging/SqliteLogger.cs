namespace API.Logging;

using API.Features.Chat.Common;
using Microsoft.Data.Sqlite;
using System.Security.Cryptography;
using System.Text;

public sealed class SqliteLogger : ILogger, IDisposable
{
    private readonly SqliteConnection _connection;

    public SqliteLogger(string connectionString)
    {
        _connection = new SqliteConnection(connectionString);
        _connection.Open();

        using var cmd = _connection.CreateCommand();
        cmd.CommandText = @"
            PRAGMA journal_mode=WAL;
            PRAGMA synchronous=NORMAL;
        ";
        cmd.ExecuteNonQuery();
    }

    public IDisposable? BeginScope<TState>(TState state) where TState : notnull => null;
    public bool IsEnabled(LogLevel logLevel) => true;

    public void Log<TState>(
        LogLevel logLevel,
        EventId eventId,
        TState state,
        Exception? exception,
        Func<TState, Exception?, string> formatter)
    {
#if DEBUG
        try
        {
            var dict = new Dictionary<string, object?>(StringComparer.OrdinalIgnoreCase);

            if (state is IEnumerable<KeyValuePair<string, object?>> logValues)
            {
                foreach (var kvp in logValues)
                    dict[kvp.Key] = kvp.Value;
            }

            string? traceId = dict.GetValueOrDefault("TraceIdentifier")?.ToString();

            using var cmd = _connection.CreateCommand();

            string message = formatter(state, exception);

            // error logs
            if (logLevel >= LogLevel.Error || exception != null)
            {
                var stackTrace = exception?.StackTrace ?? string.Empty;
                var stackTraceHash = string.IsNullOrWhiteSpace(stackTrace)
                    ? string.Empty
                    : Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(stackTrace)));

                if (!string.IsNullOrWhiteSpace(stackTraceHash))
                {
                    using var stackCommand = _connection.CreateCommand();
                    stackCommand.CommandText = @"
                        INSERT INTO StackTraces (StackTraceHash, StackTraceText)
                        VALUES (@hash, @text)
                        ON CONFLICT(StackTraceHash) DO UPDATE SET
                            LastSeenUtc = CURRENT_TIMESTAMP,
                            SeenCount = SeenCount + 1;";
                    stackCommand.Parameters.AddWithValue("@hash", stackTraceHash);
                    stackCommand.Parameters.AddWithValue("@text", stackTrace);
                    stackCommand.ExecuteNonQuery();
                }

                cmd.CommandText = @"
                    INSERT INTO ErrorLogs (TraceId, ErrorMethod, ExceptionType, Message, StackTrace, StackTraceHash, ExceptionSource, TargetSite, Endpoint)
                    VALUES (@tid, @method, @type, @msg, @stack, @hash, @source, @target, @endpoint)";
                cmd.Parameters.AddWithValue("@tid", traceId ?? "N/A");
                cmd.Parameters.AddWithValue("@method", dict.GetValueOrDefault("Method")?.ToString() ?? string.Empty);
                cmd.Parameters.AddWithValue("@type", exception?.GetType().Name ?? "Manual Error");
                cmd.Parameters.AddWithValue("@msg", exception?.Message ?? formatter(state, exception));
                cmd.Parameters.AddWithValue("@stack", stackTrace);
                cmd.Parameters.AddWithValue("@hash", string.IsNullOrWhiteSpace(stackTraceHash) ? DBNull.Value : stackTraceHash);
                cmd.Parameters.AddWithValue("@source", exception?.Source ?? string.Empty);
                cmd.Parameters.AddWithValue("@target", exception?.TargetSite?.ToString() ?? string.Empty);
                cmd.Parameters.AddWithValue("@endpoint", dict.GetValueOrDefault("Path")?.ToString() ?? string.Empty);
            }
            // query logs
            else if (traceId != null && dict.TryGetValue("CommandText", out var query))
            {
                cmd.CommandText = @"
                    INSERT INTO QueryLogs (TraceId, QueryText, ElapsedMS)
                    VALUES (@tid, @q, @e)";
                cmd.Parameters.AddWithValue("@tid", traceId);
                cmd.Parameters.AddWithValue("@q", query?.ToString());
                cmd.Parameters.AddWithValue("@e", dict.GetValueOrDefault("ElapsedMilliseconds") ?? 0);
            }
            // request logs
            else if (traceId != null && (dict.ContainsKey("Method") || dict.ContainsKey("Path")))
            {
                cmd.CommandText = @"
                    INSERT INTO Requests (Endpoint, Method, RequestBody, StatusCode, ElapsedMS, TraceId)
                    VALUES (@path, @meth, @body, @status, @elapsed, @tid)";
                cmd.Parameters.AddWithValue("@path", dict.GetValueOrDefault("Path") ?? "");
                cmd.Parameters.AddWithValue("@meth", dict.GetValueOrDefault("Method") ?? "");
                cmd.Parameters.AddWithValue("@body", dict.GetValueOrDefault("Body") ?? "");
                cmd.Parameters.AddWithValue("@status", dict.GetValueOrDefault("StatusCode") ?? 0);
                cmd.Parameters.AddWithValue("@elapsed", dict.GetValueOrDefault("ElapsedMilliseconds") ?? 0);
                cmd.Parameters.AddWithValue("@tid", traceId);
            }
            else
            {
                return; // nothing to log
            }

            int executed = cmd.ExecuteNonQuery();
        }
        catch (Exception ex)
        {
            Console.WriteLine($"[SqliteLogger Error] {ex.Message}");
        }
#endif
    }

    public void Dispose()
    {
        _connection.Dispose();
    }
}
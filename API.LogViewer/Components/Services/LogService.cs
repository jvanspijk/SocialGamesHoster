using Microsoft.Data.Sqlite;
using API.LogViewer.Components.Models;

namespace API.LogViewer.Components.Services;

public interface ILogService
{
        Task<List<RequestEntry>> GetLatestRequestsAsync(int limit = 50);
        Task<RequestEntry?> GetRequestByTraceIdAsync(string traceId);
        Task<List<TraceDetail>> GetTraceDetailsAsync(string traceId);
        Task<List<QuerySummary>> GetQuerySummariesAsync();
        Task<List<ExceptionEntry>> GetLatestExceptionsAsync(int limit = 200);
}

public class LogService : ILogService
{
    private readonly string _connectionString;

    public LogService(IConfiguration configuration)
    {
        var dbPath = configuration["DebugLogPath"] ?? "../API/logs/debug_logs.db";

        var builder = new SqliteConnectionStringBuilder
        {
            DataSource = dbPath,
            Mode = SqliteOpenMode.ReadOnly,
            Cache = SqliteCacheMode.Shared
        };

        _connectionString = builder.ConnectionString;
    }

    public async Task<List<RequestEntry>> GetLatestRequestsAsync(int limit = 50)
    {
        var requests = new List<RequestEntry>();

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        const string sql = @"
            SELECT Timestamp, Method, Endpoint, RequestBody, StatusCode, ElapsedMS, TraceId 
            FROM Requests 
            ORDER BY ID DESC 
            LIMIT @limit";

        using var cmd = new SqliteCommand(sql, conn);
        cmd.Parameters.AddWithValue("@limit", limit);

        using var reader = await cmd.ExecuteReaderAsync();
        while (await reader.ReadAsync())
        {
            requests.Add(new RequestEntry(
                reader.GetString(0),
                reader.GetString(1),
                reader.GetString(2),
                reader.IsDBNull(3) ? string.Empty : reader.GetString(3),
                reader.GetInt32(4),
                reader.GetDouble(5),
                reader.GetString(6)
            ));
        }

        return requests;
    }

    public async Task<List<TraceDetail>> GetTraceDetailsAsync(string traceId)
    {
        // Guard clause for invalid input
        if (string.IsNullOrWhiteSpace(traceId)) return new();

        var details = new List<TraceDetail>();

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        const string sql = @"
            SELECT 'SQL' as Type, QueryText, ElapsedMS, Timestamp FROM QueryLogs WHERE TraceId = @tid
            UNION ALL
            SELECT 'EXCEPTION' as Type, Message || CHAR(10) || StackTrace, 0, Timestamp FROM Exceptions WHERE TraceId = @tid
            ORDER BY Timestamp ASC";

        using var cmd = new SqliteCommand(sql, conn);
        cmd.Parameters.AddWithValue("@tid", traceId);

        using var reader = await cmd.ExecuteReaderAsync();
        while (await reader.ReadAsync())
        {
            details.Add(new TraceDetail(
                reader.GetString(0),
                reader.GetString(1),
                reader.GetDouble(2),
                reader.GetString(3)
            ));
        }

        return details;
    }

    public async Task<RequestEntry?> GetRequestByTraceIdAsync(string traceId)
    {
        if (string.IsNullOrWhiteSpace(traceId))
        {
            return null;
        }

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        const string sql = @"
            SELECT Timestamp, Method, Endpoint, RequestBody, StatusCode, ElapsedMS, TraceId
            FROM Requests
            WHERE TraceId = @tid
            ORDER BY ID DESC
            LIMIT 1";

        using var cmd = new SqliteCommand(sql, conn);
        cmd.Parameters.AddWithValue("@tid", traceId);

        using var reader = await cmd.ExecuteReaderAsync();
        if (!await reader.ReadAsync())
        {
            return null;
        }

        return new RequestEntry(
            reader.IsDBNull(0) ? string.Empty : reader.GetString(0),
            reader.IsDBNull(1) ? string.Empty : reader.GetString(1),
            reader.IsDBNull(2) ? string.Empty : reader.GetString(2),
            reader.IsDBNull(3) ? string.Empty : reader.GetString(3),
            reader.IsDBNull(4) ? 0 : reader.GetInt32(4),
            reader.IsDBNull(5) ? 0 : reader.GetDouble(5),
            reader.IsDBNull(6) ? string.Empty : reader.GetString(6)
        );
    }

    public async Task<List<QuerySummary>> GetQuerySummariesAsync()
    {
        var queryLogs = new List<QueryLogProjection>();

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        const string sql = @"
            SELECT q.QueryText,
                   q.ElapsedMS,
                   q.Timestamp,
                   COALESCE(r.Endpoint, '') AS Endpoint
            FROM QueryLogs q
            LEFT JOIN Requests r ON r.ID = (
                SELECT MAX(r2.ID)
                FROM Requests r2
                WHERE r2.TraceId = q.TraceId
            )
            WHERE QueryText IS NOT NULL AND TRIM(QueryText) <> ''
            ORDER BY q.Timestamp DESC";

        using var cmd = new SqliteCommand(sql, conn);
        using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            queryLogs.Add(new QueryLogProjection(
                reader.GetString(0),
                reader.IsDBNull(1) ? 0 : reader.GetDouble(1),
                reader.IsDBNull(2) ? string.Empty : reader.GetString(2),
                reader.IsDBNull(3) ? string.Empty : reader.GetString(3)
            ));
        }

        return queryLogs
            .GroupBy(log => log.QueryText, StringComparer.Ordinal)
            .Select(group =>
            {
                var endpoints = group
                    .Select(log => NormalizeEndpoint(log.Endpoint))
                    .Distinct(StringComparer.OrdinalIgnoreCase)
                    .OrderBy(endpoint => endpoint, StringComparer.OrdinalIgnoreCase)
                    .ToList();

                return new QuerySummary(
                    group.Key,
                    group.Average(log => log.ElapsedMs),
                    group.Count(),
                    group.Max(log => log.Timestamp),
                    endpoints);
            })
            .OrderByDescending(summary => summary.LatestTimestamp, StringComparer.OrdinalIgnoreCase)
            .ToList();
    }

    private static string NormalizeEndpoint(string endpoint)
    {
        if (string.IsNullOrWhiteSpace(endpoint))
        {
            return "Unknown endpoint";
        }

        var path = endpoint;
        var querySeparatorIndex = path.IndexOf('?');
        if (querySeparatorIndex >= 0)
        {
            path = path[..querySeparatorIndex];
        }

        var segments = path
            .Split('/', StringSplitOptions.RemoveEmptyEntries)
            .Select(segment => long.TryParse(segment, out _) || Guid.TryParse(segment, out _) ? "{id}" : segment)
            .ToArray();

        return segments.Length == 0 ? "/" : $"/{string.Join('/', segments)}";
    }

    private sealed record QueryLogProjection(string QueryText, double ElapsedMs, string Timestamp, string Endpoint);

    public async Task<List<ExceptionEntry>> GetLatestExceptionsAsync(int limit = 200)
    {
        var exceptions = new List<ExceptionEntry>();

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        try
        {
            const string sql = @"
                SELECT ID, Timestamp, TraceId, ErrorMethod, ExceptionType, Message, StackTrace, StackTraceHash, ExceptionSource, TargetSite, Endpoint
                FROM Exceptions
                ORDER BY ID DESC
                LIMIT @limit";

            using var cmd = new SqliteCommand(sql, conn);
            cmd.Parameters.AddWithValue("@limit", limit);

            using var reader = await cmd.ExecuteReaderAsync();
            while (await reader.ReadAsync())
            {
                exceptions.Add(new ExceptionEntry(
                    reader.GetInt64(0),
                    reader.IsDBNull(1) ? string.Empty : reader.GetString(1),
                    reader.IsDBNull(2) ? string.Empty : reader.GetString(2),
                    reader.IsDBNull(3) ? string.Empty : reader.GetString(3),
                    reader.IsDBNull(4) ? "UnknownException" : reader.GetString(4),
                    reader.IsDBNull(5) ? string.Empty : reader.GetString(5),
                    reader.IsDBNull(6) ? string.Empty : reader.GetString(6),
                    reader.IsDBNull(7) ? string.Empty : reader.GetString(7),
                    reader.IsDBNull(8) ? string.Empty : reader.GetString(8),
                    reader.IsDBNull(9) ? string.Empty : reader.GetString(9),
                    reader.IsDBNull(10) ? string.Empty : reader.GetString(10)
                ));
            }
        }
        catch (SqliteException ex) when (ex.Message.Contains("no such column", StringComparison.OrdinalIgnoreCase))
        {
            const string legacySql = @"
                SELECT ID, Timestamp, TraceId, ExceptionType, Message, StackTrace, Endpoint
                FROM Exceptions
                ORDER BY ID DESC
                LIMIT @limit";

            using var legacyCmd = new SqliteCommand(legacySql, conn);
            legacyCmd.Parameters.AddWithValue("@limit", limit);

            using var legacyReader = await legacyCmd.ExecuteReaderAsync();
            while (await legacyReader.ReadAsync())
            {
                exceptions.Add(new ExceptionEntry(
                    legacyReader.GetInt64(0),
                    legacyReader.IsDBNull(1) ? string.Empty : legacyReader.GetString(1),
                    legacyReader.IsDBNull(2) ? string.Empty : legacyReader.GetString(2),
                    string.Empty,
                    legacyReader.IsDBNull(3) ? "UnknownException" : legacyReader.GetString(3),
                    legacyReader.IsDBNull(4) ? string.Empty : legacyReader.GetString(4),
                    legacyReader.IsDBNull(5) ? string.Empty : legacyReader.GetString(5),
                    string.Empty,
                    string.Empty,
                    string.Empty,
                    legacyReader.IsDBNull(6) ? string.Empty : legacyReader.GetString(6)
                ));
            }
        }

        return exceptions;
    }
}


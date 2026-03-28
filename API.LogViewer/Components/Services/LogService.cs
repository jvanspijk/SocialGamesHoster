using Microsoft.Data.Sqlite;
using API.LogViewer.Components.Models;

namespace API.LogViewer.Components.Services;

public interface ILogService
{
        Task<List<RequestEntry>> GetLatestRequestsAsync(int limit = 50);
        Task<List<TraceDetail>> GetTraceDetailsAsync(string traceId);
        Task<List<QuerySummary>> GetQuerySummariesAsync();
        Task<List<ErrorEntry>> GetLatestErrorsAsync(int limit = 200);
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
            SELECT 'ERROR' as Type, Message || CHAR(10) || StackTrace, 0, Timestamp FROM ErrorLogs WHERE TraceId = @tid
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

    public async Task<List<QuerySummary>> GetQuerySummariesAsync()
    {
        var summaries = new List<QuerySummary>();

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        const string sql = @"
            SELECT QueryText, AVG(ElapsedMS) AS AvgElapsedMs, COUNT(*) AS ExecutionCount
            FROM QueryLogs
            WHERE QueryText IS NOT NULL AND TRIM(QueryText) <> ''
            GROUP BY QueryText
            ORDER BY AvgElapsedMs DESC";

        using var cmd = new SqliteCommand(sql, conn);
        using var reader = await cmd.ExecuteReaderAsync();

        while (await reader.ReadAsync())
        {
            summaries.Add(new QuerySummary(
                reader.GetString(0),
                reader.IsDBNull(1) ? 0 : reader.GetDouble(1),
                reader.IsDBNull(2) ? 0 : reader.GetInt32(2)
            ));
        }

        return summaries;
    }

    public async Task<List<ErrorEntry>> GetLatestErrorsAsync(int limit = 200)
    {
        var errors = new List<ErrorEntry>();

        using var conn = new SqliteConnection(_connectionString);
        await conn.OpenAsync();

        const string sql = @"
            SELECT ID, Timestamp, TraceId, ExceptionType, Message, StackTrace, Endpoint
            FROM ErrorLogs
            ORDER BY ID DESC
            LIMIT @limit";

        using var cmd = new SqliteCommand(sql, conn);
        cmd.Parameters.AddWithValue("@limit", limit);

        using var reader = await cmd.ExecuteReaderAsync();
        while (await reader.ReadAsync())
        {
            errors.Add(new ErrorEntry(
                reader.GetInt64(0),
                reader.IsDBNull(1) ? string.Empty : reader.GetString(1),
                reader.IsDBNull(2) ? string.Empty : reader.GetString(2),
                reader.IsDBNull(3) ? "UnknownException" : reader.GetString(3),
                reader.IsDBNull(4) ? string.Empty : reader.GetString(4),
                reader.IsDBNull(5) ? string.Empty : reader.GetString(5),
                reader.IsDBNull(6) ? string.Empty : reader.GetString(6)
            ));
        }

        return errors;
    }
}


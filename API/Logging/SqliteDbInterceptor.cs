namespace API.Logging;

using Microsoft.EntityFrameworkCore.Diagnostics;
using System.Data.Common;

public class SqliteDbInterceptor(ILogger<SqliteDbInterceptor> logger, IHttpContextAccessor httpContextAccessor) : DbCommandInterceptor
{
    private readonly IHttpContextAccessor _http = httpContextAccessor;
    public override async ValueTask<DbDataReader> ReaderExecutedAsync(
        DbCommand command,
        CommandExecutedEventData eventData,
        DbDataReader result,
        CancellationToken cancellationToken = default)
    {
        LogCommand(command, eventData);
        return await base.ReaderExecutedAsync(command, eventData, result, cancellationToken);
    }

    public override async ValueTask<int> NonQueryExecutedAsync(
        DbCommand command,
        CommandExecutedEventData eventData,
        int result,
        CancellationToken cancellationToken = default)
    {
        LogCommand(command, eventData);
        return await base.NonQueryExecutedAsync(command, eventData, result, cancellationToken);
    }

    private void LogCommand(DbCommand command, CommandExecutedEventData eventData)
    {
#if DEBUG
        string? traceId = _http.HttpContext?.TraceIdentifier;

        logger.LogInformation(
                "EF Query Executed. TraceIdentifier: {TraceIdentifier}, CommandText: {CommandText}, ElapsedMilliseconds: {ElapsedMilliseconds}",
            traceId,
            command.CommandText,
            eventData.Duration.TotalMilliseconds);
#endif
    }   
}
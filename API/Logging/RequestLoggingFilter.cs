namespace API.Logging;

using System.Diagnostics;
using Microsoft.AspNetCore.Http;

public sealed class RequestLoggingFilter(ILogger<RequestLoggingFilter> logger) : IEndpointFilter
{
    private const int MaxBodyLength = 2048;
    internal const string LoggedMethodItemKey = "DebugLog.Method";
    internal const string LoggedPathItemKey = "DebugLog.Path";
    internal const string LoggedBodyItemKey = "DebugLog.Body";
    private readonly ILogger<RequestLoggingFilter> _logger = logger;

    public async ValueTask<object?> InvokeAsync(
        EndpointFilterInvocationContext context,
        EndpointFilterDelegate next)
    {
        var http = context.HttpContext;

        var method = http.Request.Method;
        var path = http.Request.Path.ToString();
        var traceId = http.TraceIdentifier;

        string? body = null;

        if (http.Request.ContentLength is > 0 and < MaxBodyLength)
        {
            http.Request.EnableBuffering();

            using var reader = new StreamReader(http.Request.Body, leaveOpen: true);
            body = await reader.ReadToEndAsync();

            http.Request.Body.Position = 0;
        }

        http.Items[LoggedMethodItemKey] = method;
        http.Items[LoggedPathItemKey] = path;
        http.Items[LoggedBodyItemKey] = body ?? string.Empty;

        var sw = Stopwatch.StartNew();

        try
        {
            return await next(context);
        }
        finally
        {
            sw.Stop();

            var statusCode = http.Response?.StatusCode ?? 0;

            _logger.LogInformation(
                "RequestLog {Method} {Path} {StatusCode} {ElapsedMilliseconds} {TraceIdentifier} {Body}",
                method,
                path,
                statusCode,
                sw.ElapsedMilliseconds,
                traceId,
                body ?? string.Empty
            );
        }
    }
}

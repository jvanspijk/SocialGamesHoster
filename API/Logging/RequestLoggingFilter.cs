namespace API.Logging;

using System.Diagnostics;
using Microsoft.AspNetCore.Http;

public sealed class RequestLoggingFilter(ILogger<RequestLoggingFilter> logger) : IEndpointFilter
{
    private const int MaxBodyLength = 2048;
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

        var sw = Stopwatch.StartNew();

        object? result = null;
        Exception? exception = null;

        try
        {
            result = await next(context);
            return result;
        }
        catch (Exception ex)
        {
            exception = ex;
            throw;
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

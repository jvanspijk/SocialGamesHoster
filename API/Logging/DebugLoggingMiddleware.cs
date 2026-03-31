using System.Diagnostics;

namespace API.Logging;

public class DebugLoggingMiddleware(RequestDelegate next, ILogger<DebugLoggingMiddleware> logger)
{
    private static readonly PathString ApiPrefix = new("/api");
    private const int MaxBodyLength = 2048;

    public async Task InvokeAsync(HttpContext context)
    {
        if (!context.Request.Path.StartsWithSegments(ApiPrefix))
        {
            await next(context);
            return;
        }

        string method = context.Request.Method;
        string path = context.Request.Path.ToString();
        string traceId = context.TraceIdentifier;
        string body = string.Empty;

        if (context.Request.ContentLength is > 0 and < MaxBodyLength)
        {
            context.Request.EnableBuffering();
            using var reader = new StreamReader(context.Request.Body, leaveOpen: true);
            body = await reader.ReadToEndAsync();
            context.Request.Body.Position = 0;
        }

        var sw = Stopwatch.StartNew();

        try
        {
            await next(context);

            sw.Stop();
            int statusCode = context.Response.StatusCode;

            logger.LogInformation(
                "RequestLog {Method} {Path} {StatusCode} {ElapsedMilliseconds} {TraceIdentifier} {Body}",
                method,
                path,
                statusCode,
                sw.ElapsedMilliseconds,
                traceId,
                body
            );
        }
        catch (Exception ex)
        {
            sw.Stop();

            logger.LogError(ex,
                "Unhandled exception {Method} {Path} {TraceIdentifier} {Body}",
                method,
                path,
                traceId,
                body);

            throw; // Re-throw to allow standard error handling (e.g., ProblemDetails)
        }
    }
}
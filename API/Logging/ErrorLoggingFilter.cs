namespace API.Logging;

public class ErrorLoggingFilter(ILogger<ErrorLoggingFilter> logger) : IEndpointFilter
{
    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        try
        {
            return await next(context);
        }
        catch (Exception ex)
        {
#if DEBUG
            var method = context.HttpContext.Items.TryGetValue(RequestLoggingFilter.LoggedMethodItemKey, out var methodValue)
                ? methodValue?.ToString() ?? context.HttpContext.Request.Method
                : context.HttpContext.Request.Method;
            var path = context.HttpContext.Items.TryGetValue(RequestLoggingFilter.LoggedPathItemKey, out var pathValue)
                ? pathValue?.ToString() ?? context.HttpContext.Request.Path.ToString()
                : context.HttpContext.Request.Path.ToString();
            var body = context.HttpContext.Items.TryGetValue(RequestLoggingFilter.LoggedBodyItemKey, out var bodyValue)
                ? bodyValue?.ToString() ?? string.Empty
                : string.Empty;

            logger.LogError(ex,
                "Unhandled exception {Method} {Path} {TraceIdentifier} {Body}",
                method,
                path,
                context.HttpContext.TraceIdentifier,
                body);
#endif

            throw; // Re-throw so ASP.NET's error pages/handlers still work
        }
    }
}

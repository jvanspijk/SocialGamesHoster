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
            logger.LogError(ex,
                "An unhandled exception occurred at {Path}. Trace: {TraceIdentifier}",
                context.HttpContext.Request.Path,
                context.HttpContext.TraceIdentifier);
#endif

            throw; // Re-throw so ASP.NET's error pages/handlers still work
        }
    }
}

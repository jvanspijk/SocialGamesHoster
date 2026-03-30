using Microsoft.AspNetCore.Mvc;

namespace API.Filters;

public sealed class HttpErrorFormattingFilter : IEndpointFilter
{
    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        try
        {
            var result = await next(context);

            if (result is not IStatusCodeHttpResult { StatusCode: >= 300 } statusCodeResult)
            {
                return result;
            }

            int statusCode = statusCodeResult.StatusCode.Value;

            if (result is IValueHttpResult valueHttpResult && valueHttpResult.Value is ProblemDetails)
            {
                return result;
            }

            string? detail = null;
            if (result is IValueHttpResult rawValueResult && rawValueResult.Value is string stringValue)
            {
                detail = stringValue;
            }

            return Results.Json(new
            {
                status = statusCode,
                title = GetDefaultTitle(statusCode),
                detail
            }, statusCode: statusCode);
        }
        catch (Exception)
        {
            return Results.Json(new
            {
                status = StatusCodes.Status500InternalServerError,
                title = "Internal Server Error",
                detail = "An unexpected error occurred."
            }, statusCode: StatusCodes.Status500InternalServerError);
        }
    }

    private static string GetDefaultTitle(int statusCode) => statusCode switch
    {
        StatusCodes.Status400BadRequest => "Bad Request",
        StatusCodes.Status401Unauthorized => "Unauthorized",
        StatusCodes.Status403Forbidden => "Forbidden",
        StatusCodes.Status404NotFound => "Not Found",
        StatusCodes.Status409Conflict => "Conflict",
        StatusCodes.Status422UnprocessableEntity => "Unprocessable Entity",
        StatusCodes.Status500InternalServerError => "Internal Server Error",
        _ => "An error occurred"
    };
}
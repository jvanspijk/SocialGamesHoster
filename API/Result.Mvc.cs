using API.Validation;
using Microsoft.AspNetCore.Mvc;

namespace API;

/// <summary>
/// Extension methods to convert Result<T> to IActionResult or IResult for ASP.NET Core MVC and Minimal APIs.
/// </summary>
public static class ResultMvcExtensions
{
    public static IActionResult AsActionResult(this Error error)
    {
        return new ObjectResult(error.Message) { StatusCode = (int)error.StatusCode };
    }
    public static IActionResult AsActionResult<T>(
        this Result<T> result,
        Func<T, IActionResult>? onSuccess = null,
        Func<Error, IActionResult>? onFailure = null
    ) where T : notnull
    {
        return result switch
        {
            Failure<T> failure => onFailure != null
                ? onFailure(failure.Error)
                : new ObjectResult(failure.Error.Message) { StatusCode = (int)failure.Error.StatusCode },
            Success<T> success => onSuccess != null
                ? onSuccess(success.Value)
                : new OkObjectResult(success.Value),
            _ => throw new InvalidOperationException("Unrecognized Result type"),
        };
    }

    public static IResult AsIResult<T>(this Result<T> result) where T : notnull
    {
        return result switch
        {
            Failure<T> failure => Results.Problem(failure.Error.Message, statusCode: (int)failure.Error.StatusCode),
            Success<T> success => Results.Ok(success.Value),
            _ => throw new InvalidOperationException("Unrecognized Result type"),
        };
    }
}




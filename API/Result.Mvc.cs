using API.Domain;
using API.Domain.Validation;
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
        onSuccess ??= val => new OkObjectResult(val);
        onFailure ??= err => new ObjectResult(err.Message) { StatusCode = (int)err.StatusCode };
        return result.Match(onSuccess, onFailure);        
    }

    public static IResult AsIResult<T>(this Result<T> result) where T : notnull
    {
        return result.Match(
            val => Results.Ok(val), 
            err => Results.Problem(err.Message, statusCode: (int)err.StatusCode)
        );        
    }
}




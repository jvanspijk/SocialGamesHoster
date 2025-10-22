using API.Domain;
using API.Domain.Validation;
using Microsoft.AspNetCore.Mvc;

namespace API;

/// <summary>
/// Extension methods to convert Result<T> to IActionResult or IResult for ASP.NET Core MVC and Minimal APIs.
/// </summary>
public static class ResultMvcExtensions
{
    public static IResult AsIResult(this Error error)
    {
        return Results.Problem(error.Message, statusCode: (int)error.StatusCode);
    }
    public static IResult AsIResult<T>(this Result<T> result) where T : notnull
    {
        return result.Match(
            val => Results.Ok(val), 
            err => Results.Problem(err.Message, statusCode: (int)err.StatusCode)
        );        
    }
}




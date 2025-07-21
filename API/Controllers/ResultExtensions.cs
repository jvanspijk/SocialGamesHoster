using API.Validation;
using LanguageExt.Common;
using Microsoft.AspNetCore.Mvc;
using System.ComponentModel.DataAnnotations;

namespace API.Controllers;


/// <summary>
/// Because I use a functional style Result<T>
/// I use this class to limit the amount of bloat
/// necessary in the controller implementations
/// </summary>
public static class ResultExtensions
{
    /// <summary>
    /// Converts a Result<T> to an IActionResult.
    /// </summary>
    /// <typeparam name="T"></typeparam>
    /// <param name="result"></param>
    /// <returns>IActionResult</returns>
    public static IActionResult ToActionResult<T>(this Result<T> result)
    {
        return result.Match<IActionResult>(
            value => new OkObjectResult(value),
            ex =>
            {
                var errorDetails = new
                {
                    Message = "An unexpected error occurred.",
                    ExceptionType = ex.GetType().FullName,
                    ExceptionMessage = ex.Message,
                    StackTrace = ex.StackTrace // We should be careful with exposing stack traces in production
                };

                if (ex is CustomException || ex is ValidationException)
                {
                    return new BadRequestObjectResult(errorDetails);
                }               

                return new ObjectResult(errorDetails)
                {
                    StatusCode = 500
                };
            }
        );
    }
    /// <summary>
    /// convert a Result<T> to an object of type T.
    /// It should only be used when you are sure that the Result is successful.
    /// This is to reduce nesting.
    /// </summary>
    /// <typeparam name="T"></typeparam>
    /// <param name="result"></param>
    /// <returns>T</returns>
    public static T ToObjectUnsafe<T>(this Result<T> result)
    {
        return result.Match(
            Succ: value => value,
            Fail: ex => throw ex
        );
    }
}

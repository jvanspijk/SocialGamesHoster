using API.Validation;
using LanguageExt.Common;
using Microsoft.AspNetCore.Mvc;
using System.ComponentModel.DataAnnotations;

namespace API.Controllers;

// Because I use a functional style Result<T>
// I use this class to limit the amount of bloat
// necessary in the controller implementations
public static class ControllerExtensions
{
    public static IActionResult ToActionResult<T>(this Result<T> result)
    {
        return result.Match<IActionResult>(
            value => new OkObjectResult(value),
            ex =>
            {
                if (ex is CustomException || ex is ValidationException)
                {
                    return new BadRequestObjectResult(ex.Message);
                }
                return new StatusCodeResult(500);
            }
        );
    }
}

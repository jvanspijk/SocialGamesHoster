using API.Validation;
using Microsoft.AspNetCore.Mvc;

namespace API;
public readonly partial record struct Result<T>
{
    public readonly T Value => (T)_value!;
    public readonly IActionResult AsActionResult()
    {
        return _value switch
        {
            T _ => new OkObjectResult(_value),
            ValidationError err => new BadRequestObjectResult(err.Message()),
            Exception ex => new ObjectResult("An unexpected server error occured.")
            {
                StatusCode = 500,
                Value = ex.Message,
            },
            null => new NotFoundObjectResult("Requested object not found."),
            _ => new ObjectResult("An unexpected server error occured.")
            {
                StatusCode = 500,
                Value = "The result object contained an unhandled type.",
            },
        };
    }
}

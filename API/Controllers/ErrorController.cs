using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

namespace API.Controllers;

[ApiController]
[ApiExplorerSettings(IgnoreApi = true)]
public class ErrorController : ControllerBase
{
    [Route("/Error")]
    public IActionResult HandleError()
    {
        var exceptionFeature = HttpContext.Features.Get<IExceptionHandlerFeature>();
        if (exceptionFeature != null)
        {
            var exception = exceptionFeature.Error;
            // TODO: maybe logging if necessary
            return Problem(
                detail: exception.Message,
                title: "An error occurred.",
                statusCode: StatusCodes.Status500InternalServerError
            );
        }
        return Problem(
            detail: "An unexpected error occurred.",
            title: "Unknown Error",
            statusCode: StatusCodes.Status500InternalServerError
        );
    }
}

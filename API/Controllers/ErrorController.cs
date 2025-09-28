using Microsoft.AspNetCore.Diagnostics;
using Microsoft.AspNetCore.Mvc;

namespace API.Controllers;

[ApiController]
[ApiExplorerSettings(IgnoreApi = true)]
public class ErrorController(ILogger<ErrorController> logger) : ControllerBase
{
    private readonly ILogger _logger = logger;
    private static readonly EventId UnknownErrorEvent = new(1, "UnknownError");

    [Route("/Error")]
    public IActionResult HandleError()
    {
        var exceptionFeature = HttpContext.Features.Get<IExceptionHandlerFeature>();
        if (exceptionFeature != null)
        {
            
            var exception = exceptionFeature.Error;
            _logger.LogError(exception, "Operation failed: {ErrorMessage}", exception.Message);
            return Problem(
                detail: exception.Message,
                title: "An error occurred.",
                statusCode: StatusCodes.Status500InternalServerError
            );
        }
        _logger.LogError(UnknownErrorEvent, "An unexpected error occurred.");
        return Problem(
            detail: "An unexpected error occurred.",
            title: "Unknown Error",
            statusCode: StatusCodes.Status500InternalServerError
        );
    }
}

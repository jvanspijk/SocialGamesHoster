using Microsoft.AspNetCore.Mvc;
using System.Net;

namespace API.Validation;

public enum ErrorType
{
    Authentication_MissingClaims,
    Authentication_NotAuthorized,
    Authentication_Required,
    Authentication_Failed,

    UserAlreadyLoggedIn,        
    UserNotOnline, 
    NotFound,

    Validation_EmptyName,
    Validation_InvalidGuid
}

public readonly record struct APIError(ErrorType Type, string Message, HttpStatusCode StatusCode, string DebugMessage = "")
{
    public IActionResult AsActionResult()
    {
        var errorPayload = new
        {
            error = $"{Type} error: {Message}"
        };
       
        return (int)StatusCode switch
        {
            // Standard Client Errors
            400 => new BadRequestObjectResult(errorPayload),
            401 => new UnauthorizedObjectResult(errorPayload),
            403 => new ForbidResult(),
            404 => new NotFoundObjectResult(errorPayload),
            409 => new ConflictObjectResult(errorPayload),
            422 => new UnprocessableEntityObjectResult(errorPayload),

            // Standard Server Errors
            500 => new ObjectResult(errorPayload) { StatusCode = 500 },
            503 => new StatusCodeResult(503), // Service Unavailable

            // Fallback for any other custom or uncategorized code
            _ => new ObjectResult(errorPayload) { StatusCode = (int)StatusCode }
        };
    }
}

public static partial class Errors
{
    // --- AUTHENTICATION ERRORS ---
    public static APIError AuthenticationRequired(string message = "Authentication is required to access this resource.") =>
        new(ErrorType.Authentication_Required, message, HttpStatusCode.Unauthorized);

    public static APIError AuthenticationFailed(string message = "Authentication failed due to invalid credentials.") =>
        new(ErrorType.Authentication_Failed, message, HttpStatusCode.Unauthorized);

    public static APIError NotAuthorized(string message = "You are authenticated but lack the necessary permissions to perform this action.") =>
        new(ErrorType.Authentication_NotAuthorized, message, HttpStatusCode.Forbidden);

    public static APIError MissingClaims(string message = "The user token is valid but is missing required claims.") =>
        new(ErrorType.Authentication_MissingClaims, message, HttpStatusCode.Forbidden);

    // ------------
    public static APIError UserAlreadyLoggedIn(string message = "User is already logged in.") =>
        new(ErrorType.UserAlreadyLoggedIn, message, HttpStatusCode.Conflict);

    public static APIError UserNotOnline(string message = "User is not online and cannot log out.") =>
        new(ErrorType.UserNotOnline, message, HttpStatusCode.Conflict);

    public static APIError ResourceNotFound(string resourceName, string id) =>
        new(ErrorType.NotFound, $"{resourceName} with ID '{id}' was not found.", HttpStatusCode.NotFound);

    public static APIError ResourceNotFound(string message = "Resource not found.") =>
        new(ErrorType.NotFound, message, HttpStatusCode.NotFound);

    // -----------
    public static APIError EmptyName(string message = "Provided name is empty.") =>
        new(ErrorType.Validation_EmptyName, message, HttpStatusCode.BadRequest);

    public static APIError InvalidGuid(string message = "Provided GUID is invalid.") =>
        new(ErrorType.Validation_InvalidGuid, message, HttpStatusCode.BadRequest);
}
using System.Net;

namespace API.Domain.Validation;

public enum ErrorType
{
    Authentication_MissingClaims,
    Authentication_NotAuthorized,
    Authentication_Required,
    Authentication_Failed,

    UserAlreadyLoggedIn,        
    UserNotOnline, 
    NotFound,
    FailedToCreate,

    Validation,
}

public record class Error(ErrorType Type, string Message, HttpStatusCode StatusCode, string DebugMessage = "");
public record class ValidationError(string Field, string Message) : Error(ErrorType.Validation, Message, HttpStatusCode.BadRequest);
public static partial class Errors
{
    // --- AUTHENTICATION ERRORS ---
    public static Error AuthenticationRequired(string message = "Authentication is required to access this resource.") =>
        new(ErrorType.Authentication_Required, message, HttpStatusCode.Unauthorized);

    public static Error AuthenticationFailed(string message = "Authentication failed due to invalid credentials.") =>
        new(ErrorType.Authentication_Failed, message, HttpStatusCode.Unauthorized);

    public static Error Unauthorized(string message = "You are authenticated but lack the necessary permissions to perform this action.") =>
        new(ErrorType.Authentication_NotAuthorized, message, HttpStatusCode.Forbidden);

    public static Error MissingClaims(string message = "The user token is valid but is missing required claims.") =>
        new(ErrorType.Authentication_MissingClaims, message, HttpStatusCode.Forbidden);

    // ------------
    public static Error UserAlreadyLoggedIn(string message = "User is already logged in.") =>
        new(ErrorType.UserAlreadyLoggedIn, message, HttpStatusCode.Conflict);

    public static Error UserNotOnline(string message = "User is not online and cannot log out.") =>
        new(ErrorType.UserNotOnline, message, HttpStatusCode.Conflict);

    public static Error ResourceNotFound(string resourceName, int id) =>
        new(ErrorType.NotFound, $"{resourceName} with ID '{id}' was not found.", HttpStatusCode.NotFound);

    public static Error ResourceNotFound(string message = "Resource not found.") =>
        new(ErrorType.NotFound, message, HttpStatusCode.NotFound);

    // -----------    

    public static Error FailedToCreate(string resourceName, string? name = null) =>
        new(ErrorType.FailedToCreate, $"Failed to create {resourceName}{(name is not null ? $" with name {name}" : string.Empty)}.", HttpStatusCode.BadRequest);
}

public static class ValidationExtensions
{
    public static bool IsValid(this IEnumerable<ValidationError> errors)
    {        
        return !errors.Any();
    }   
    public static bool HasErrors(this IEnumerable<ValidationError> errors)
    {
        return errors.Any();
    }
}
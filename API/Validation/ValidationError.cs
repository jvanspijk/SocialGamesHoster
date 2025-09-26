namespace API.Validation;
public enum ValidationError
{
    UserAlreadyLoggedInError,
    UserNotOnlineException,
    NotFoundException       
}

public static class ValidationErrorExtensions
{        
    public static string Message(this ValidationError error)
    {
        return error switch
        {
            // User-related errors
            ValidationError.UserAlreadyLoggedInError => "The specified user is already logged in.",
            ValidationError.UserNotOnlineException => "The requested user is currently offline.",

            // Resource errors
            ValidationError.NotFoundException => "The requested resource could not be found.",

            // Default/Unknown
            _ => "An unclassified or unknown error occurred."
        };
    }
}

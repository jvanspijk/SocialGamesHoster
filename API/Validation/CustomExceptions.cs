namespace API.Validation;

// Custom exceptions that are presented to the user 
// when something goes wrong.
public class CustomException : Exception
{
    public CustomException() : base() { }
    public CustomException(string message) : base(message) { }

    public CustomException(string message, Exception innerException)
        : base(message, innerException) { }
}

public class UserAlreadyLoggedInException : CustomException
{
    public UserAlreadyLoggedInException() : base("User is already logged in.") { }

    public UserAlreadyLoggedInException(string message) : base(message) { }

    public UserAlreadyLoggedInException(string message, Exception innerException)
        : base(message, innerException) { }
}

public class UserNotOnlineException : CustomException
{
    public UserNotOnlineException() : base("User is not online and cannot log out.") { }

    public UserNotOnlineException(string message) : base(message) { }

    public UserNotOnlineException(string message, Exception innerException)
        : base(message, innerException) { }
}

public class NotFoundException : CustomException
{
    public NotFoundException() : base("Couldn't find value for given input.") { }
    public NotFoundException(string message) : base(message) { }
    public NotFoundException(string message, Exception innerException)
    : base(message, innerException) { }
}

public class EmptyNameException : CustomException
{
    public EmptyNameException() : base("Provided name is empty.") { }
    public EmptyNameException(string message) : base(message) { }
    public EmptyNameException(string message, Exception innerException)
    : base(message, innerException) { }
}

public class InvalidGuidException : CustomException
{
    public InvalidGuidException() : base("Provided GUID is invalid.") { }
    public InvalidGuidException(string message) : base(message) { }
    public InvalidGuidException(string message, Exception innerException)
    : base(message, innerException) { }
}


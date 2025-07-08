namespace API.Models;

public class UserAlreadyLoggedInException : Exception
{
    public UserAlreadyLoggedInException() : base("User is already logged in.") { }

    public UserAlreadyLoggedInException(string message) : base(message) { }

    public UserAlreadyLoggedInException(string message, Exception innerException)
        : base(message, innerException) { }
}

public class UserNotOnlineException : Exception
{
    public UserNotOnlineException() : base("User is not online and cannot log out.") { }

    public UserNotOnlineException(string message) : base(message) { }

    public UserNotOnlineException(string message, Exception innerException)
        : base(message, innerException) { }
}

public class UserDoesNotExistException : Exception
{
    public UserDoesNotExistException() : base("User does not exist.") { }
    public UserDoesNotExistException(string message) : base(message) { }
    public UserDoesNotExistException(string message, Exception innerException)
    : base(message, innerException) { }
}


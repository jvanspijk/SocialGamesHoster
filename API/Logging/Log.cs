using Microsoft.Extensions.Logging;

namespace API.Logging;

public static partial class Log
{
    // The [LoggerMessage] attribute tells the source generator what to create.
    [LoggerMessage(
        EventId = 100, // Unique ID for this log event
        Level = LogLevel.Error,
        Message = "An unhandled {ExceptionType} occurred: {ExceptionMessage}")]

    // The generated method MUST be static and partial, and MUST take ILogger as 'this'.
    // The parameters (logger, exceptionType, exceptionMessage, ex) will populate the log.
    public static partial void LogExceptionDetails(
        this ILogger logger,
        string exceptionType,
        string exceptionMessage,
        Exception ex);
}

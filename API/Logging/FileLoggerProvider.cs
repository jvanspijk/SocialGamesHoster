using Microsoft.Extensions.Logging;

namespace API.Logging;
public class FileLoggerProvider(string path, int maxFiles, LogLevel minLogLevel, LogLevel maxLogLevel) : ILoggerProvider
{
    private readonly string _path = path;
    private readonly int _maxFiles = maxFiles;
    private readonly LogLevel _minLogLevel = minLogLevel;
    private readonly LogLevel _maxLogLevel = maxLogLevel;

    public ILogger CreateLogger(string categoryName)
    {
        return new FileLogger(_path, _maxFiles, _minLogLevel, _maxLogLevel);
    }

    public void Dispose() { }
}

public static class FileLoggerExtensions
{
    public static ILoggingBuilder AddFileLogger(this ILoggingBuilder builder, 
        string filePath, int maxFilesToKeep, LogLevel minLogLevel, LogLevel maxLogLevel)
    {
        builder.AddProvider(new FileLoggerProvider(filePath, maxFilesToKeep, minLogLevel, maxLogLevel));
        return builder;
    }
}
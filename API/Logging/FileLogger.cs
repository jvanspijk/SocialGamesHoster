using System.Text;

namespace API.Logging;

/// <summary>
/// Simple rolling file logger.
/// </summary>
public class FileLogger : ILogger
{
    private readonly LogLevel _minLogLevel;
    private readonly LogLevel _maxLogLevel;
    private readonly string _logFilePath;
    private readonly int _maxFileCount;
    private readonly object _lock = new();

    private const long MaxFileSizeInBytes = 8 * 1024 * 1024; // 8 MB

    public FileLogger(string logFilePath, int maxFileCount, LogLevel minLogLevel, LogLevel maxLogLevel = LogLevel.Critical)
    {
        _logFilePath = logFilePath;
        _maxFileCount = maxFileCount;

        Directory.CreateDirectory(Path.GetDirectoryName(logFilePath)!);
        _minLogLevel = minLogLevel;
        _maxLogLevel = maxLogLevel;
    }

    public IDisposable? BeginScope<TState>(TState state) where TState : notnull => default;

    public bool IsEnabled(LogLevel logLevel)
    {
        return logLevel >= _minLogLevel && logLevel <= _maxLogLevel;
    }

    public void Log<TState>(LogLevel logLevel, EventId eventId, TState state, Exception? exception, Func<TState, Exception?, string> formatter)
    {
        if (!IsEnabled(logLevel))
        {
            return;
        }

        var message = formatter(state, exception);
        var logEntry = $"{DateTime.Now:yyyy-MM-dd HH:mm:ss.fff} [{logLevel}] - {message}{Environment.NewLine}";

        if (exception != null)
        {
            logEntry += $"Exception: {exception}{Environment.NewLine}";
        }

        lock (_lock)
        {
            PerformRollingCheck();
            File.AppendAllText(_logFilePath, logEntry);
        }
    }

    private void PerformRollingCheck()
    {
        FileInfo fileInfo = new(_logFilePath);

        if (fileInfo.Exists && fileInfo.Length >= MaxFileSizeInBytes)
        {
            string newFileName = $"{Path.GetFileNameWithoutExtension(_logFilePath)}_{DateTime.Now:yyyyMMdd_HHmmss}{Path.GetExtension(_logFilePath)}";
            string newPath = Path.Combine(Path.GetDirectoryName(_logFilePath)!, newFileName);

            File.Move(_logFilePath, newPath);

            EnforceRetention();
        }
    }

    private void EnforceRetention()
    {
        string? directoryPath = Path.GetDirectoryName(_logFilePath);
        if (directoryPath == null) return;

        var logFiles = Directory.GetFiles(directoryPath, "*.log")
                              .Select(f => new FileInfo(f))
                              .OrderByDescending(f => f.CreationTime)
                              .ToList();

        for (int i = _maxFileCount; i < logFiles.Count; i++)
        {
            logFiles[i].Delete();
        }
    }
}




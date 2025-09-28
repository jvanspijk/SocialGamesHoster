using Microsoft.Extensions.Logging;

namespace API.Logging;
public class FileLoggerProvider : ILoggerProvider
{
    private readonly string _path;
    private readonly int _maxFiles;

    public FileLoggerProvider(string path, int maxFiles)
    {
        _path = path;
        _maxFiles = maxFiles;
    }

    public ILogger CreateLogger(string categoryName)
    {
        return new FileLogger(_path, _maxFiles);
    }

    public void Dispose()
    {
    }
}

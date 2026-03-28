namespace API.LogViewer.Components.Models;

public record RequestEntry(string Timestamp, string Method, string Endpoint, string Body, int StatusCode, double ElapsedMS, string TraceId);

public record TraceDetail(string Type, string DetailText, double Duration, string Timestamp);

public record QuerySummary(string QueryText, double AverageElapsedMs, int ExecutionCount);

public record ErrorEntry(
    long Id,
    string Timestamp,
    string TraceId,
    string ExceptionType,
    string Message,
    string StackTrace,
    string Endpoint);

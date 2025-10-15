using API.Domain.Validation;

namespace API;

public static class ErrorApiExtensions
{
    public static Dictionary<string, string[]> ToProblemDetails(this IEnumerable<ValidationError> errors)
    {
        return errors
        .GroupBy(error => error.Field, StringComparer.OrdinalIgnoreCase)
        .ToDictionary(
            group => group.Key,
            group => group.Select(error => error.Message).ToArray()
        );
    }
}

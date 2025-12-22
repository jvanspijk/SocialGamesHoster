using API.Domain.Validation;

namespace API;

public static class ErrorApiExtensions
{
    extension(IEnumerable<ValidationError> errors)
    {
        public Dictionary<string, string[]> ToProblemDetails()
        {
            return errors
                .GroupBy(error => error.Field, StringComparer.OrdinalIgnoreCase)
                .ToDictionary(
                    group => group.Key,
                    group => group.Select(error => error.Message).ToArray()
                );
        }
    }
}

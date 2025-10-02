using API.Domain;

namespace API.Domain.Validation;

public interface IValidatable<T> where T : notnull
{
    Result<T> Validate(T value);
}

using API.Validation;

namespace API;

public readonly partial record struct Result<T>
{
    public Result(T value) { _value = value; }
    public Result(ValidationError validationError) { _value = validationError; }
    public Result(Exception exception) { _value = exception; }
    private readonly object? _value { get; }
    public readonly bool HasValue => _value != null && _value is T;
}

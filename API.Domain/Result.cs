using API.Domain.Validation;
using System.Diagnostics.CodeAnalysis;

namespace API.Domain;

public partial record Result
{
    public Error? Error { get; init; }
    [MemberNotNullWhen(false, nameof(Error))]
    public bool IsSuccess { get; init; }
    [MemberNotNullWhen(true, nameof(Error))]
    public bool IsFailure => !IsSuccess;

    protected internal Result(Error? error, bool isSuccess)
    {
        Error = error;
        IsSuccess = isSuccess;
    }

    public static Result Success() => new(null, true);
    public static Result Failure(Error error) => new(error, false);
    public static implicit operator Result(Error error) => Failure(error);
}

/// <summary>
/// Result type of <typeparamref name="T"/>, <see cref="Error"/>.
/// It will contain a <typeparamref name="T"/> Value or <see cref="Error"/> Error, but never both.
/// </summary>
/// <typeparam name="T"></typeparam>
public abstract partial record Result<T> : Result where T : notnull
{
    public new Error? Error { get; init; }
    public T? Value { get; init; }
    protected Result(T? value, Error? error, bool isSuccess) : base(error, isSuccess)
    {       
        Value = value;
        Error = error;
        IsSuccess = isSuccess;
    }
    public static implicit operator Result<T>(T value) => new Success<T>(value);
    public static implicit operator Result<T>(Error error) => new Failure<T>(error);    

    /// <summary>
    /// True if the result is successful.
    /// </summary>
    [MemberNotNullWhen(true, nameof(Value))]
    [MemberNotNullWhen(false, nameof(Error))]
    public new bool IsSuccess { get; init; }
    [MemberNotNullWhen(false, nameof(Value))]
    [MemberNotNullWhen(true, nameof(Error))]
    public new bool IsFailure => !IsSuccess;

    public bool TryGetValue([MaybeNullWhen(false)] out T value)
    {
        value = IsSuccess ? Value : default;
        return IsSuccess;
    }

    public bool TryGetError([MaybeNullWhen(false)] out Error error)
    {
        error = !IsSuccess ? Error : default;
        return !IsSuccess;
    }    
}


internal record Success<T>(T Value) : Result<T>(Value, null, Value != null) where T : notnull
{
    public new T Value => base.Value!; // Ensures that the linter knows Value is not null here
}
internal record Failure<T>(Error Error) : Result<T>(default, Error, false) where T : notnull 
{ 
    public new Error Error => base.Error!; // Ensures that the linter knows Error is not null here
}




using API.Models;
using API.Validation;
using System.Diagnostics.CodeAnalysis;
using static System.Runtime.InteropServices.JavaScript.JSType;

namespace API;

/// <summary>
/// Result type of <typeparamref name="T"/>, <see cref="APIError"/>.
/// It will contain a <typeparamref name="T"/> Value or <see cref="APIError"/> Error, but never both.
/// </summary>
/// <typeparam name="T"></typeparam>
public readonly partial record struct Result<T> where T : notnull
{
    [MemberNotNullWhen(true, nameof(Value))]
    [MemberNotNullWhen(false, nameof(Error))]
    public readonly bool Ok => _value != null && _value is T;
    public readonly T? Value => _value is T value ? value : default;
    public readonly APIError? Error => _value is APIError error ? error : null;

    private readonly object _value;
    public Result(T value) { _value = value; }
    public Result(APIError error) { _value = error; }

    // Implicit operator overrides so that you can return Result<T> objects as T or error
    // e,g, public Result<Player> Player => new Player();
    public static implicit operator Result<T>(T value) { return new(value); }
    public static implicit operator Result<T>(APIError error) { return new Result<T>(error); }
}
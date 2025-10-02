using API.Domain.Validation;
using System.Threading.Tasks;

namespace API.Domain;

/// <summary>
/// Chains operations on <see cref="Result{T}"/> types, enabling functional-style composition of success and failure flows.
/// </summary>
public static class ResultChainingExtensions
{
    /// <summary>
    /// Chains this result with the next operation, which returns a <see cref="Result{TOut}"/>.
    /// If this result is successful, applies <paramref name="next"/> to the value.
    /// Otherwise, propagates the failure.
    /// </summary>
    /// <typeparam name="T"></typeparam>
    /// <typeparam name="TOut"></typeparam>
    /// <param name="result"></param>
    /// <param name="next"></param>
    /// <returns></returns>
    /// <exception cref="InvalidOperationException"></exception>
    public static Result<TOut> Then<T, TOut>(this Result<T> result, Func<T, Result<TOut>> next)
        where T : notnull
        where TOut : notnull
    {
        return result switch
        {
            Success<T> success => next(success.Value),
            Failure<T> failure => new Failure<TOut>(failure.Error),
            _ => throw new InvalidOperationException("Unrecognized Result type")
        };
    }

    public static Result<TNew> Map<T, TNew>(this Result<T> result, Func<T, TNew> mapSuccess)
        where T : notnull
        where TNew : notnull
    {
        return result switch
        {
            Success<T> success => new Success<TNew>(mapSuccess(success.Value)),
            Failure<T> failure => new Failure<TNew>(failure.Error),
            _ => throw new InvalidOperationException("Unrecognized Result type")
        };
    }        

    public static TOut Match<T, TOut>(this Result<T> result, Func<T, TOut> onSuccess, Func<Error, TOut> onFailure)
        where T : notnull
    {
        return result switch
        {
            Success<T> success => onSuccess(success.Value),
            Failure<T> failure => onFailure(failure.Error),
            _ => throw new InvalidOperationException("Unrecognized Result type")
        };
    }
}

public static class ResultAsyncChainingExtensions
{
    public static async Task<Result<TOut>> Then<TIn, TOut>(
    this Task<Result<TIn>> sourceTask,
    Func<TIn, Result<TOut>> next)
    where TIn : notnull
    where TOut : notnull
    {
        Result<TIn> sourceResult = await sourceTask;
        return sourceResult.Then(next);
    }

    public static async Task<Result<TOut>> Map<TIn, TOut>(
        this Task<Result<TIn>> sourceTask,
        Func<TIn, TOut> next)
        where TIn : notnull
        where TOut : notnull
    {
        Result<TIn> sourceResult = await sourceTask;
        return sourceResult.Map(next);
    }

    public static async Task<TOut> Match<T, TOut>(
    this Task<Result<T>> sourceTask,
    Func<T, TOut> onSuccess,
    Func<Error, TOut> onFailure)
    where T : notnull
    {
        var result = await sourceTask;
        return result switch
        {
            Success<T> success => onSuccess(success.Value),
            Failure<T> failure => onFailure(failure.Error),
            _ => throw new InvalidOperationException("Unrecognized Result type")
        };
    }

    public static async Task<TOut> MatchAsync<T, TOut>(
        this Task<Result<T>> sourceTask,
        Func<T, Task<TOut>> onSuccess,
        Func<Error, Task<TOut>> onFailure)
        where T : notnull
    {
        var result = await sourceTask;
        return await (result switch
        {
            Success<T> success => onSuccess(success.Value),
            Failure<T> failure => onFailure(failure.Error),
            _ => throw new InvalidOperationException("Unrecognized Result type")
        });
    }

    public static Task<Result<TOut>> ThenAsync<TIn, TOut>(
    this Task<Result<TIn>> sourceTask,
    Func<TIn, Task<Result<TOut>>> next)
    where TIn : notnull
    where TOut : notnull

    {
        return sourceTask.MatchAsync(
            onSuccess: next,
            onFailure: error => Task.FromResult<Result<TOut>>(error)
        );
    }

    public static Task<Result<TNew>> MapAsync<TIn, TNew>(
        this Task<Result<TIn>> sourceTask,
        Func<TIn, Task<TNew>> mapSuccess)
        where TIn : notnull
        where TNew : notnull
    {
        return sourceTask.ThenAsync<TIn, TNew>(async value =>
        {
            TNew newValue = await mapSuccess(value);
            return new Success<TNew>(newValue);
        });
    }    
}
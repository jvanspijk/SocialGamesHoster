using API.Domain.Validation;
using System.Threading.Tasks;

namespace API.Domain;
#pragma warning disable CS8509 // The switch expression does not handle all possible values of its input type (it is not exhaustive).
/// <summary>
/// Chains operations on <see cref="Result{T}"/> types, enabling functional-style composition of success and failure flows.
/// </summary>
public static class ResultChainingExtensions
{
    public static Result<TOut> Bind<T, TOut>(this Result<T> result, Func<T, Result<TOut>> function)
        where T : notnull
        where TOut : notnull
    {
        return (result) switch
        {
            Failure<T> failure => failure.Error,
            Success<T> success => function(success.Value)
        };
    }

    public static Result<TNew> Map<T, TNew>(this Result<T> result, Func<T, TNew> mapSuccess)
        where T : notnull
        where TNew : notnull
    {
        return result switch
        {
            Success<T> success => new Success<TNew>(mapSuccess(success.Value)),
            Failure<T> failure => new Failure<TNew>(failure.Error)
        };
    }        

    public static TOut Match<T, TOut>(this Result<T> result, Func<T, TOut> onSuccess, Func<Error, TOut> onFailure)
        where T : notnull
    {
        return result switch
        {
            Success<T> success => onSuccess(success.Value),
            Failure<T> failure => onFailure(failure.Error)
        };
    }
}

public static class ResultAsyncChainingExtensions
{
    public static async Task<Result<TOut>> Bind<TIn, TOut>(
    this Task<Result<TIn>> sourceTask,
    Func<TIn, Result<TOut>> next)
    where TIn : notnull
    where TOut : notnull
    {
        Result<TIn> sourceResult = await sourceTask;
        return sourceResult.Bind(next);
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
        return result.Match(onSuccess, onFailure);
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
            Failure<T> failure => onFailure(failure.Error)
        });
    }      
}
#pragma warning restore CS8509 // The switch expression does not handle all possible values of its input type (it is not exhaustive).

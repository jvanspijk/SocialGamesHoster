using API.Validation;

namespace API;

public readonly partial record struct Result<T>
{
    public async Task<Result<TOut>> Bind<TOut>(Func<T, Task<Result<TOut>>> next)
        where TOut : notnull
    {
        if (Ok)
        {
            return await next(Value!);
        }

        return new Result<TOut>(Error.Value);
    }

    /// <summary>
    /// Transforms the inner value T into TOut, only if the current result is Success.
    /// Propagates Errors/Exceptions otherwise.
    /// </summary>
    public readonly Result<TNew> Map<TNew>(Func<T, TNew> mapSuccess) where TNew : notnull
    {
        if(Ok)
        {
            return new Result<TNew>(mapSuccess(Value));
        }

        return new Result<TNew>(Error.Value);       
    }
}

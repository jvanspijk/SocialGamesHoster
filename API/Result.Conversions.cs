using Microsoft.AspNetCore.Mvc;

namespace API;
public readonly partial record struct Result<T>
{
    public readonly IActionResult AsActionResult()
    {
        if(Ok)
        {
            return new OkObjectResult(Value);
        }
        return Error.Value.AsActionResult();
    }

    /// <summary>
    /// Use a mapping function for the success case.
    /// </summary>
    /// <typeparam name="TResult"></typeparam>
    /// <param name="mapSuccess"></param>
    /// <returns></returns>
    public readonly IActionResult AsActionResult<TResult>(Func<T, TResult> mapSuccess)
    {
        if(Ok)
        {
            return new OkObjectResult(mapSuccess(Value));
        }
       
        return AsActionResult();
    }
}




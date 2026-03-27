using Microsoft.AspNetCore.Mvc.Infrastructure;
using Microsoft.Extensions.Caching.Memory;
using System.Diagnostics;

namespace API.Filters;

public class CacheInvalidatorFilter(IMemoryCache cache) : IEndpointFilter
{
    private readonly IMemoryCache _cache = cache;

    private static void InvalidateAll(IMemoryCache cache)
    {
        if (cache == null) return;

        if (cache is MemoryCache concreteCache)
        {
            concreteCache.Clear();
        }
        else
        {
            Debug.Fail("Cache is not of type MemoryCache. Cannot clear cache.");
        }
    }

    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        var result = await next(context);

        var method = context.HttpContext.Request.Method;

        if (IsMutation(method) && IsSuccessful(context.HttpContext, result))
        {
            InvalidateAll(_cache);
        }

        return result;
    }

    private static bool IsMutation(string method) =>
        HttpMethods.IsPost(method) || HttpMethods.IsPut(method) ||
        HttpMethods.IsPatch(method) || HttpMethods.IsDelete(method);

    private static bool IsSuccessful(HttpContext context, object? result)
    {
        if (result is IStatusCodeActionResult statusCodeResult)
        {
            return statusCodeResult.StatusCode >= 200 && statusCodeResult.StatusCode < 300;
        }

        return context.Response.StatusCode >= 200 && context.Response.StatusCode < 300;
    }
}

/// <summary>
/// Attribute to mark requests that shouldn't clear the cache (e.g. Search)
/// </summary>
[AttributeUsage(AttributeTargets.Method)]
public class DisableCacheInvalidationAttribute : Attribute { }
using API.DataAccess;
using Microsoft.AspNetCore.Mvc.Infrastructure;
using Microsoft.EntityFrameworkCore;

namespace API;

public sealed class DebugRequireSaveChangesFilter(APIDatabaseContext dbContext) : IEndpointFilter
{
    private readonly APIDatabaseContext _dbContext = dbContext;

    public async ValueTask<object?> InvokeAsync(EndpointFilterInvocationContext context, EndpointFilterDelegate next)
    {
        object? result = await next(context);

        string method = context.HttpContext.Request.Method;

        if (IsSuccessful(context.HttpContext, result) && _dbContext.ChangeTracker.HasChanges())
        {
            string pendingChanges = string.Join(
                ", ",
                _dbContext.ChangeTracker.Entries()
                    .Where(e => e.State is EntityState.Added or EntityState.Modified or EntityState.Deleted)
                    .Select(e => $"{e.Entity.GetType().Name}:{e.State}")
            );

            throw new InvalidOperationException(
                $"Detected pending tracked changes after successful mutation request. " +
                $"Did you forget to call SaveChangesAsync on IRepository? Pending: {pendingChanges}");
        }

        return result;
    }

    private static bool IsSuccessful(HttpContext context, object? result)
    {
        if (result is IStatusCodeActionResult statusCodeResult)
        {
            return statusCodeResult.StatusCode is >= 200 and < 300;
        }

        return context.Response.StatusCode is >= 200 and < 300;
    }
}
using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    public record Response(int Id, string Name, string Description)
    : IProjectable<Role, Response>
    {
        public static Expression<Func<Role, Response>> Projection =>
            role => new Response(
                role.Id,
                role.Name,
                role.Description
            );
    }
    public static async Task<IResult> HandleAsync(RoleRepository repository, IMemoryCache cache, int ruleSetId)
    {
        string cacheKey = GetCacheKey(ruleSetId);
        List<Response>? result = await cache.GetOrCreateAsync(cacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(10);
            return await repository.GetAllFromRulesetAsync<Response>(ruleSetId);
        });

        if(result == null)
        {
            return Results.NotFound();
        }
        return Results.Ok(result);
    }

    public static void InvalidateCache(IMemoryCache cache, int ruleSetId)
    {
        string cacheKey = GetCacheKey(ruleSetId);
        cache.Remove(cacheKey);
    }

    private static string GetCacheKey(int ruleSetId) => $"{nameof(GetRoles)}_{ruleSetId}";
}

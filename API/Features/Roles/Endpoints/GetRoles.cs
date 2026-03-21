using API.DataAccess;

using API.Domain.Entities;
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
    public static async Task<IResult> HandleAsync(IRepository<Role> repository, IMemoryCache cache, int ruleSetId)
    {
        string cacheKey = GetCacheKey(ruleSetId);
        Response[] result = await cache.GetOrCreateAsync(cacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(30);
            return await repository.GetArrayReadOnlyAsync<Response>(r => r.RulesetId == ruleSetId);
        }) ?? [];

        if(result == null) // Not possible. TODO: Add the ruleset repo to check if the ruleset exists. If it doesn't, return NotFound. If it does, return an empty array.
        {
            return Results.NotFound();
        }
        return Results.Ok(result);
    }
    private static string GetCacheKey(int ruleSetId) => $"{nameof(GetRoles)}_{ruleSetId}";
}

using API.DataAccess;

using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    private static string GetCacheKey(int ruleSetId) => $"{nameof(GetRoles)}_{ruleSetId}";
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
    public static async Task<Results<Ok<Response[]>, ProblemHttpResult>> HandleAsync(IRepository<Role> repository, IRepository<Ruleset> rulesetRepository, IMemoryCache cache, int ruleSetId)
    {
        if (!await rulesetRepository.ExistsAsync(r => r.Id == ruleSetId))
        {
            return APIResults.NotFound<Ruleset>(ruleSetId);
        }

        string cacheKey = GetCacheKey(ruleSetId);
        Response[] result = await cache.GetOrCreateAsync(cacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(30);
            return await repository.GetArrayReadOnlyAsync<Response>(r => r.RulesetId == ruleSetId);
        }) ?? [];

        return APIResults.Ok(result);
    }
}

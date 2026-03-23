using API.DataAccess;

using API.Domain.Entities;
using API.Features.Rulesets.Common;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetFullRuleset
{
    private static string CacheKey(int rulesetId) => $"{nameof(GetFullRuleset)}_{rulesetId}";
    public record Response(int Id, string Name, string Description, List<AbilityInfo> Abilities, List<RoleInfo> Roles)
    : IProjectable<Ruleset, Response>
    {
        public static Expression<Func<Ruleset, Response>> Projection =>
            rs => new Response(
                    rs.Id, rs.Name, rs.Description,
                    rs.Abilities.Select(a => new AbilityInfo(a.Id, a.Name, a.Description)).ToList(),
                    rs.Roles.Select(
                        r => new RoleInfo(
                            r.Id, r.Name, r.Description,
                            r.Abilities.Select(a => a.Id).ToList(),
                            r.KnowsAbout.Where(k => k.KnowledgeType == KnowledgeType.Role).Select(k => k.TargetId).ToList()
                        )).ToList()
                    );

    }
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<Ruleset> repository, IMemoryCache cache, int rulesetId)
    {
        Response? response = await cache.GetOrCreateAsync(CacheKey(rulesetId), async entry =>
        {
            entry.SlidingExpiration = TimeSpan.FromMinutes(15);
            return await repository.GetReadOnlyAsync<Response>(r => r.Id == rulesetId, splitQuery: true);
        });
        if (response == null)
        {
            return APIResults.NotFound<Ruleset>(rulesetId);
        }
        return APIResults.Ok(response);
    }
}

using API.DataAccess;
using API.Domain.Entities;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetRulesetAbilities
{
    private static string CacheKey(int rulesetId) => $"{nameof(GetRulesetAbilities)}_{rulesetId}";

    public record Response(int Id, string Name, string Description) : IProjectable<Ability, Response>
    {
        public static Expression<Func<Ability, Response>> Projection =>
            ability => new Response(ability.Id, ability.Name, ability.Description);
    }

    public static async Task<IResult> HandleAsync(
        IRepository<Ability> repository,
        IRepository<Ruleset> rulesetRepo,
        IMemoryCache cache,
        int rulesetId)
    {
        var result = await cache.GetOrCreateAsync(CacheKey(rulesetId), async entry =>
        {
            entry.SlidingExpiration = TimeSpan.FromMinutes(30);
            bool exists = await rulesetRepo.CountAsync(r => r.Id == rulesetId) > 0;
            if (!exists) return null;
            return await repository.GetArrayReadOnlyAsync<Response>(a => a.RulesetId == rulesetId);
        });

        if (result == null)
        {
            return Results.NotFound($"Ruleset {rulesetId} does not exist.");
        }

        return Results.Ok(result);
    }
}

using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetRuleset
{
    public record Response(int Id, string Name, string Description)
        : IProjectable<Ruleset, Response>
    {
        public static Expression<Func<Ruleset, Response>> Projection =>
            rs => new Response(rs.Id, rs.Name, rs.Description);
    }

    public static async Task<IResult> HandleAsync(RulesetRepository repository, IMemoryCache cache, int rulesetId)
    {
        string cacheKey = $"{nameof(GetRuleset)}{rulesetId}";
        Response? response = await cache.GetOrCreateAsync(cacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(10);
            return await repository.GetAsync<Response>(rulesetId);
        });
        if (response == null)
        {
            return Results.Problem($"Ruleset with id {rulesetId} not found.");
        }
        return Results.Ok(response);
    }
}

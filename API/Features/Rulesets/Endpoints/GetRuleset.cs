using API.DataAccess;
using API.Domain.Entities;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetRuleset
{
    private static string CacheKey(int rulesetId) => $"{nameof(GetRuleset)}_{rulesetId}";
    public record Response(int Id, string Name, string Description)
        : IProjectable<Ruleset, Response>
    {
        public static Expression<Func<Ruleset, Response>> Projection =>
            rs => new Response(rs.Id, rs.Name, rs.Description);
    }

    public static async Task<IResult> HandleAsync(IRepository<Ruleset> repository, IMemoryCache cache, int rulesetId)
    {
        Response? response = await cache.GetOrCreateAsync(CacheKey(rulesetId), async entry =>
        {
            entry.SlidingExpiration = TimeSpan.FromMinutes(5);
            return await repository.GetReadOnlyAsync<Response>(r=> r.Id == rulesetId);
        });

        if (response == null)
        {
            return Results.Problem($"Ruleset with id {rulesetId} not found.");
        }

        return Results.Ok(response);
    }
}

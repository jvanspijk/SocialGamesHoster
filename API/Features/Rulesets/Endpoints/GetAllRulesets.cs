using API.DataAccess;


using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetAllRulesets
{
    private static readonly string CacheKey = "GetAllRulesetsResponse";
    public record Response(int Id, string Name, string Description)
        : IProjectable<Ruleset, Response>
    {
        public static Expression<Func<Ruleset, Response>> Projection =>
            rs => new Response(rs.Id, rs.Name, rs.Description);
    }
    public static async Task<Ok<Response[]>> HandleAsync(IRepository<Ruleset> repository, IMemoryCache cache)
    {
        Response[]? response = await cache.GetOrCreateAsync(CacheKey, async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(5);
            return await repository.GetArrayReadOnlyAsync<Response>();
        }) ?? [];
        return APIResults.Ok(response);
    }
}

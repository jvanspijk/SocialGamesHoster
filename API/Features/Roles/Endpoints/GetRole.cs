using API.DataAccess;
using API.Domain.Entities;
using API.Features.Roles.Common;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public class GetRole
{
    private static string GetCacheKey(int roleId) => $"{nameof(Role)}_{roleId}";
    public record Response(int Id, string Name, string Description, List<AbilityInfo> Abilities)
        : IProjectable<Role, Response>
    {
        public static Expression<Func<Role, Response>> Projection =>
            role => new Response(
                role.Id,
                role.Name,
                role.Description,
                role.Abilities.Select(a => new AbilityInfo(a.Id, a.Name, a.Description)).ToList()
            );
    }
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<Role> repository, IMemoryCache cache, int id)
    {
        Response? result = await cache.GetOrCreateAsync(GetCacheKey(id), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(10);
            return await repository.GetReadOnlyAsync<Response>(r => r.Id == id);
        });

        if (result == null)
        {
            return APIResults.NotFound<Role>(id);
        }
        return APIResults.Ok(result);
    }
}

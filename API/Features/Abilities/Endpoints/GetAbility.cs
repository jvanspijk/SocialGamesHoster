using API.DataAccess;
using API.Domain.Entities;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Abilities.Endpoints;

public class GetAbility
{
    private static string CacheKey(int id) => $"{nameof(GetAbility)}_{id}";
    public record Response(int Id, string Name, string Description) : IProjectable<Ability, Response>
    {
        public static Expression<Func<Ability, Response>> Projection =>
            ability => new Response(ability.Id, ability.Name, ability.Description);
    }
    public static async Task<IResult> HandleAsync(
        IRepository<Ability> repository,
        IMemoryCache cache,
        int id)
    {
        Response? result = await cache.GetOrCreateAsync(
            CacheKey(id),
            async entry => {
                entry.SlidingExpiration = TimeSpan.FromMinutes(10);
                return await repository.GetReadOnlyAsync<Response>(a => a.Id == id);
            } 
        );

        if (result == null)
        {
            return Results.NotFound($"Ability with id {id} not found.");
        }

        return Results.Ok(result);
    }
}

using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class GetGamePlayers
{
    private static string CacheKey(int gameId) => $"{nameof(GetGamePlayers)}_{gameId}";
    public record Response(int Id, string Name) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(player.Id, player.Name);
    }
    public static async Task<Results<Ok<Response[]>, ProblemHttpResult>> HandleAsync(IRepository<Player> repository, IMemoryCache cache, int gameId)
    {
        Response[] result = await cache.GetOrCreateAsync(CacheKey(gameId), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(5);
            return await repository.GetArrayReadOnlyAsync<Response>(p => p.GameId == gameId);
        }) ?? [];

        return APIResults.Ok(result);
    }
}

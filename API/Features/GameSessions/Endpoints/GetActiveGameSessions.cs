using API.DataAccess;
using API.Domain.Models;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;
public static class GetActiveGameSessions
{
    private static string CacheKey() => $"{nameof(GetActiveGameSessions)}_ActiveGames";
    public record Response(int Id, int RulesetId, string Status, int CurrentRoundNumber) 
        : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.RulesetId,
                gs.Status.ToFriendlyString(),
                gs.RoundNumber
            );
    }
    public static async Task<IResult> HandleAsync(IRepository<GameSession> repository, IMemoryCache cache)     
    {
        var activeGames = await cache.GetOrCreateAsync(CacheKey(), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromSeconds(30);
            return await repository.GetArrayReadOnlyAsync<Response>(gs => gs.IsActive);
        }) ?? [];

        return Results.Ok(activeGames); 
    }

}

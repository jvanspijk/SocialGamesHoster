using API.DataAccess;
using API.Domain.Models;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class GetGameSession
{
    private static string CacheKey(int gameId) => $"{nameof(GetGameSession)}_{gameId}";
    public record Response(int Id, int RulesetId, string Status) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.RulesetId,
                gs.Status.ToFriendlyString()
            );
    }
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, IMemoryCache cache, int gameId)
    {
        Response? response = await cache.GetOrCreateAsync(CacheKey(gameId), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(5);
            return await repository.GetReadOnlyAsync<Response>(g => g.Id == gameId);
        });

        if (response == null)
        {
            return APIResults.NotFound<GameSession>(gameId);
        }

        return APIResults.Ok(response);
    }
}

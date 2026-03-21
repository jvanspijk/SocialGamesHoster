using API.DataAccess;
using API.Domain.Models;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class GetAllGameSessions
{
    private static string CacheKey() => $"{nameof(GetAllGameSessions)}";
    public record Response(int Id, string RulesetName, List<int> ParticipantIds, string Status) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.Ruleset!.Name,
                gs.Participants.Select(p => p.Id).ToList(),
                gs.Status.ToFriendlyString()
            );
    }

    public static async Task<IResult> HandleAsync(
        IRepository<GameSession> repository,
        IMemoryCache cache)
    {
        var gameSessions = await cache.GetOrCreateAsync(CacheKey(), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(1);
            return await repository.GetArrayReadOnlyAsync<Response>();
        });
        return Results.Ok(gameSessions);
    }


}

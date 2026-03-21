using API.DataAccess;
using API.Domain.Entities;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using API.Features.GameSessions.Hubs;
using Microsoft.AspNetCore.SignalR;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;
using System.Text;

namespace API.Features.GameSessions.Endpoints;

public static class StartNewRound
{
    private static string CacheKey(int rulesetId) => $"{nameof(StartNewRound)}_{rulesetId}";
    public record Request(int GameId, int NewPhaseId);
    public record Response(int Id, int CurrentRound, DateTimeOffset StartedAt, int? RoundPhaseId)
        : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gamesession => new Response(gamesession.Id, gamesession.RoundNumber, gamesession.RoundStartedAt, gamesession.CurrentPhaseId);
    }

    public static async Task<IResult> HandleAsync(IRepository<GameSession> gameRepository, IRepository<RoundPhase> phaseRepository, IHubContext<GameSessionsHub, IGameSessionsHub> hub, IMemoryCache cache, Request request)
    {
        GameSession? session = await gameRepository.GetWithTrackingAsync(request.GameId);
        if (session == null)
        {
            return Results.NotFound($"Game with id {request.GameId} not found.");
        }

        RoundPhase[] allPhases = await cache.GetOrCreateAsync(CacheKey(session.RulesetId), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromDays(1);
            return await phaseRepository.GetArrayReadOnlyAsync(rp => rp.RulesetId == session.RulesetId);
        }) ?? [];
            
        bool validPhase = allPhases.Any(phase => phase.Id == request.NewPhaseId);
        if (!validPhase)
        {
            return Results.BadRequest($"Phase with id {request.NewPhaseId} is not valid for the ruleset of the game.");
        }

        session.RoundNumber += 1;
        session.CurrentPhaseId = request.NewPhaseId;
        session.RoundStartedAt = DateTimeOffset.UtcNow;

        await GameSessionsHub.NotifyRoundStarted(hub, session.Id);
        Response response = new(session.Id, session.RoundNumber, session.RoundStartedAt, session.CurrentPhaseId);
        return Results.Ok(response);
    }
}

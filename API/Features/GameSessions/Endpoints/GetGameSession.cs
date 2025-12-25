using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;
using API.Features.GameSessions.Common;

namespace API.Features.GameSessions.Endpoints;

public static class GetGameSession
{
    public record Response(int Id, int RulesetId, string Status) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.RulesetId,
                gs.Status.ToFriendlyString()
            );
    }
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, int gameId)
    {        
        var gameSession = await repository.GetAsync<Response>(gameId);
        if (gameSession == null)
        {
            return Results.NotFound($"Game session with ID {gameId} not found.");
        }
        return Results.Ok(gameSession);
    }
}

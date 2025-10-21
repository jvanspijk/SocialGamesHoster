using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;
public static class GetActiveGameSessions
{
    public record Response(int Id, int RulesetId, string Status, int CurrentRoundNumber) 
        : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.RulesetId,
                gs.Status.ToString(),
                gs.Rounds.Count(r => r.StartedTime.HasValue)
            );
    }
    public static async Task<IResult> HandleAsync(GameSessionRepository repository)     
    {
        var activeGames = await repository.GetAllActiveAsync<Response>();
        if (activeGames.Count == 0)
        {
            return Results.NotFound("No active game sessions found.");
        }
        return Results.Ok(activeGames); 
    }

}

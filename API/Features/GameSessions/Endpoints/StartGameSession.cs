using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class StartGameSession
{
    public record Response(int Id, string Status) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.Status.ToString()
            );
    }
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, int gameId)
    {
        Result<GameSession> result = await repository.StartGameSession(gameId);
        if(result.IsFailure)
        {
            return result.AsIResult();
        }
        GameSession session = result.Value;
        var response = new Response(session.Id, session.Status.ToString());
        return Results.Ok(response);
    }
}

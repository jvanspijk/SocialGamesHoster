using API.DataAccess;
using API.Domain.Models;
using Microsoft.AspNetCore.Http.HttpResults;
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
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, int gameId)
    {
        GameSession? session = await repository.GetWithTrackingAsync(gameId);
        if(session == null)
        {
            return APIResults.NotFound($"Game session with id `{gameId}` not found.");
        }

        // TODO: check old logic for starting game sessions in the old repo
        session.Status = GameStatus.Running;

        await repository.SaveChangesAsync();

        var response = new Response(session.Id, session.Status.ToString());
        return APIResults.Ok(response);
    }
}

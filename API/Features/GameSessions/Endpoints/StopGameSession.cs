using API.DataAccess;
using API.Domain;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class StopGameSession
{
    public record Response(int Id, string Status) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.Status.ToString()
            );
    }
    public static async Task<IResult> HandleAsync(IRepository<GameSession> repository, RoundTimer timer, int gameId)
    {
        GameSession? session = await repository.GetWithTrackingAsync(gameId);
        if (session == null)
        {
            return Results.NotFound($"Game session with id `{gameId}` not found.");
        }

        // TODO: check old logic for starting game sessions in the old repo
        session.Status = GameStatus.Finished;

        await repository.SaveChangesAsync();

        var response = new Response(session.Id, session.Status.ToString());
        return Results.Ok(response);
    }
}
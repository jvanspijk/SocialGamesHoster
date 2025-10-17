using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;

namespace API.Features.GameSessions.Endpoints;

public class EndGameSession
{
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, int gameId)
    {
        Result<GameSession> result = await repository.EndGameSession(gameId);
        if(!result.IsSuccess)
        {
            return result.AsIResult();
        }

        return Results.NoContent();
    }
}

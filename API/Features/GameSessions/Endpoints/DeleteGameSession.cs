using API.DataAccess.Repositories;
using API.Domain;

namespace API.Features.GameSessions.Endpoints;

public static class DeleteGameSession
{
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, int gameId)
    {
        Result result = await repository.DeleteAsync(gameId);
        if(result.IsFailure)
        {
            return Results.Problem(result.Error?.Message);
        }
        return Results.NoContent();
    }
}

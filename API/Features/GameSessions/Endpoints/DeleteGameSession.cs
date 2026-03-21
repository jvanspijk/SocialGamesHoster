using API.DataAccess;
using API.Domain;
using API.Domain.Models;

namespace API.Features.GameSessions.Endpoints;

public static class DeleteGameSession
{
    public static async Task<IResult> HandleAsync(IRepository<GameSession> repository, int gameId)
    {
        var gameSession = await repository.GetWithTrackingAsync(gameId);
        if(gameSession == null)
        {
            return Results.NotFound($"Game session with id {gameId} not found.");
        }
        repository.Remove(gameSession);
        await repository.SaveChangesAsync();
        return Results.NoContent();
    }
}

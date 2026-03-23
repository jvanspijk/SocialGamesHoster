using API.DataAccess;
using API.Domain.Models;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.GameSessions.Endpoints;

public static class DeleteGameSession
{
    public static async Task<Results<NoContent, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, int gameId)
    {
        var gameSession = await repository.GetWithTrackingAsync(gameId);
        if(gameSession == null)
        {
            return APIResults.NotFound<GameSession>(gameId);
        }
        repository.Remove(gameSession);
        await repository.SaveChangesAsync();
        return APIResults.NoContent();
    }
}

using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Features.GameSessions.Responses;

namespace API.Features.GameSessions.Endpoints;

public static class StartGameSession
{
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, int gameId)
    {
        Result<GameSession> result = await repository.StartGameSession(gameId);
        if(result.IsFailure)
        {
            return result.AsIResult();
        }
        GameSession session = result.Value;
        var response = new ActiveGameResponse(session.Id, session.RulesetId, session.Status.ToString(), 0);
        return Results.Ok(response);
    }
}

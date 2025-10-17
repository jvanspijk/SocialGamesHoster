using API.DataAccess.Repositories;
using API.Features.GameSessions.Responses;

namespace API.Features.GameSessions.Endpoints;
public static partial class GetActiveGameSessions
{

    public static async Task<IResult> HandleAsync(GameSessionRepository repository)     
    {
        var activeGames = await repository.GetAllActiveAsync<ActiveGameResponse>();
        if (activeGames.Count == 0)
        {
            return Results.NotFound("No active game sessions found.");
        }
        return Results.Ok(activeGames); 
    }

}

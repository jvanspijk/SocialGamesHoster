using API.DataAccess.Repositories;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayersFromGame
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository, int gameId)
    {
        List<PlayerNameResponse> result = await repository.GetAllFromGameAsync<PlayerNameResponse>(gameId);
        return Results.Ok(result);
    }
}

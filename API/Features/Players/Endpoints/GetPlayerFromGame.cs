using API.DataAccess.Repositories;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayerFromGame
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository, string name, int gameId)
    {
        PlayerResponse? result = await repository.GetByNameAsync<PlayerResponse>(name, gameId);
        if (result == null)
        {
            return Results.NotFound($"Player with name '{name}' not found.");
        }
        return Results.Ok(result);
    }
}

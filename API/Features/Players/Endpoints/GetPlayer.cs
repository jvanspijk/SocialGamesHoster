using API.DataAccess.Repositories;
using API.Features.Players;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayer
{
    public static async Task<IResult> HandleByIdAsync(PlayerRepository repository, int id)
    {
        PlayerResponse? result = await repository.GetAsync<PlayerResponse>(id);
        if (result == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
        return Results.Ok(result);
    }

    public static async Task<IResult> HandleByNameAsync(PlayerRepository repository, string name)
    {
        PlayerResponse? result = await repository.GetByNameAsync<PlayerResponse>(name);
        if (result == null)
        {
            return Results.NotFound($"Player with name '{name}' not found.");
        }
        return Results.Ok(result);
    }
}

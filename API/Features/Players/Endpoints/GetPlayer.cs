using API.DataAccess.Repositories;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayer
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository, int id)
    {
        PlayerResponse? result = await repository.GetAsync<PlayerResponse>(id);
        if (result == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
        return Results.Ok(result);
    }
}

using API.DataAccess.Repositories;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayers
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository)
    {
        List<PlayerResponse> result = await repository.GetAllAsync<PlayerResponse>();
        return Results.Ok(result);
    }
}

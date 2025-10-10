using API.DataAccess.Repositories;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayers
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository)
    {
        List<PlayerNameResponse> result = await repository.GetAllAsync<PlayerNameResponse>();
        return Results.Ok(result);
    }
}

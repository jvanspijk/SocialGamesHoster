using API.AdminFeatures.Players.Responses;
using API.DataAccess.Repositories;

namespace API.AdminFeatures.Players.Endpoints;

public static class GetPlayers
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository)
    {
        List<FullPlayerResponse> result = await repository.GetAllAsync<FullPlayerResponse>();
        return Results.Ok(result);
    }
}
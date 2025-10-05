using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class GetPlayers
{
    public static async Task<IResult> HandleAsync(PlayerService playerService)
    {
        var result = await playerService.GetAllAsync();
        return result.AsIResult();
    }
}

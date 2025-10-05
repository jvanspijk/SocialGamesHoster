using API.Features.Players;

namespace API.Features.Players.Endpoints;

public static class GetPlayer
{
    public static async Task<IResult> HandleByIdAsync(PlayerService playerService, int id)
    {
        var result = await playerService.GetAsync(id);
        return result.AsIResult();
    }

    public static async Task<IResult> HandleByNameAsync(PlayerService playerService, string name)
    {
        var result = await playerService.GetByNameAsync(name);
        return result.AsIResult();
    }
}

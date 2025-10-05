using API.Domain;
using API.Features.Players.Responses;

namespace API.AdminFeatures.Players.Endpoints;

public static class GetPlayer
{
    public static async Task<IResult> HandleByIdAsync(AdminPlayerService playerService, int id)
    {
        var result = await playerService.GetAsync(id);
        return result.AsIResult();
    }

    public static async Task<IResult> HandleByNameAsync(AdminPlayerService playerService, string name)
    {
        var result = await playerService.GetByNameAsync(name);
        return result.AsIResult();
    }
}

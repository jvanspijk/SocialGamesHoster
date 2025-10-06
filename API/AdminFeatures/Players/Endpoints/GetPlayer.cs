using API.AdminFeatures.Players.Responses;
using API.DataAccess.Repositories;

namespace API.AdminFeatures.Players.Endpoints;

public static class GetPlayer
{
    public static async Task<IResult> HandleByIdAsync(PlayerRepository repository, int id)
    {
        FullPlayerResponse? result = await repository.GetAsync<FullPlayerResponse>(id);
        if(result == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
        return Results.Ok(result);
    }

    public static async Task<IResult> HandleByNameAsync(PlayerRepository repository, string name)
    {
        FullPlayerResponse? result = await repository.GetByNameAsync<FullPlayerResponse>(name);
        if (result == null)
        {
            return Results.NotFound($"Player with name '{name}' not found.");
        }
        return Results.Ok(result);
    }    
}

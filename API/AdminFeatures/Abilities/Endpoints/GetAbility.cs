using API.AdminFeatures.Abilities.Responses;
using API.DataAccess.Repositories;

namespace API.AdminFeatures.Abilities.Endpoints;

public static class GetAbility
{
    public static async Task<IResult> HandleAsync(AbilityRepository repository, int id)
    {
        AbilityResponse? result = await repository.GetAsync<AbilityResponse>(id);
        if (result == null)
        {
            return Results.NotFound($"Ability with id {id} not found.");
        }
        return Results.Ok(result);
    }
}

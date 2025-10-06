using API.DataAccess.Repositories;
using API.Features.Abilities.Responses;

namespace API.Features.Abilities.Endpoints;

public class GetAbility
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

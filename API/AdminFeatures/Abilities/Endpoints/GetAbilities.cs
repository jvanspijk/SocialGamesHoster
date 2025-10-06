using API.AdminFeatures.Abilities.Responses;
using API.DataAccess.Repositories;

namespace API.AdminFeatures.Abilities.Endpoints;

public static class GetAbilities
{
    public static async Task<IResult> HandleAsync(AbilityRepository repository)
    {
        List<AbilityResponse> result = await repository.GetAllAsync<AbilityResponse>();
        return Results.Ok(result);
    }
}

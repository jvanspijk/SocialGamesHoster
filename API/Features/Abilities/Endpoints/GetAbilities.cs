using API.DataAccess.Repositories;
using API.Features.Abilities.Responses;
using API.Features.Players;

namespace API.Features.Abilities.Endpoints;

public class GetAbilities
{
    public static async Task<IResult> HandleAsync(AbilityRepository repository)
    {
        List<AbilityResponse> result = await repository.GetAllAsync<AbilityResponse>();
        return Results.Ok(result);
    }
}

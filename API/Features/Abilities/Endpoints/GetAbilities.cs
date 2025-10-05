using API.Features.Players;

namespace API.Features.Abilities.Endpoints;

public class GetAbilities
{
    public static async Task<IResult> HandleAsync(AbilityService abilityService)
    {
        var result = await abilityService.GetAllAsync();
        return result.AsIResult();
    }
}

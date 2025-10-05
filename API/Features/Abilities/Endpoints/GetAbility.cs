namespace API.Features.Abilities.Endpoints;

public class GetAbility
{
    public static async Task<IResult> HandleAsync(AbilityService abilityService, int id)
    {
        var result = await abilityService.GetAsync(id);
        return result.AsIResult();
    }
}

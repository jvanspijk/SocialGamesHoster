namespace API.AdminFeatures.Abilities.Endpoints;

public static class GetAbility
{
    public static async Task<IResult> HandleAsync(AdminAbilityService abilityService, int id)
    {
        var result = await abilityService.GetAsync(id);
        return result.AsIResult();
    }
}

namespace API.AdminFeatures.Abilities.Endpoints;

public static class GetAbilities
{
    public static async Task<IResult> HandleAsync(AdminAbilityService abilityService)
    {
        var result = await abilityService.GetAllAsync();
        return result.AsIResult();
    }
}

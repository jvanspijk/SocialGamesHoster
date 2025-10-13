using API.DataAccess.Repositories;
using API.Features.Abilities.Responses;

namespace API.Features.Abilities.Endpoints;

public static class GetAbilitiesFromRuleset
{
    public static async Task<IResult> HandleAsync(AbilityRepository repository, int rulesetId)
    {
        List<AbilityResponse> result = await repository.GetAllFromRulesetAsync<AbilityResponse>(rulesetId);
        return Results.Ok(result);
    }
}

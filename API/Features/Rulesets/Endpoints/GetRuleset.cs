using API.DataAccess;
using API.DataAccess.Repositories;
using API.Features.Abilities.Responses;
using API.Features.Rulesets.Responses;

namespace API.Features.Rulesets.Endpoints;

public static class GetRuleset
{
    public static async Task<IResult> HandleAsync(RulesetRepository repository, int rulesetId)
    {
        RulesetResponse? response = await repository.GetAsync<RulesetResponse>(rulesetId);
        if (response == null)
        {
            return Results.Problem($"Ruleset with id {rulesetId} not found.");
        }
        return Results.Ok(response);
    }
}

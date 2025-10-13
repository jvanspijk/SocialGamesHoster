using API.DataAccess.Repositories;
using API.Features.Rulesets.Responses;

namespace API.Features.Rulesets.Endpoints;

public static class GetRuleset
{
    public static async Task<IResult> HandleAsync(RulesetRepository repository, int id)
    {
        RulesetResponse? response = await repository.GetAsync<RulesetResponse>(id);
        if (response == null)
        {
            return Results.Problem($"Ruleset with id {id} not found.");
        }
        return Results.Ok(response);
    }
}

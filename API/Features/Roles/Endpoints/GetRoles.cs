using API.DataAccess.Repositories;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    public static async Task<IResult> HandleAsync(RoleRepository repository, int ruleSetId)
    {
        List<RoleWithoutAbilitiesResponse> result = await repository.GetAllFromRulesetAsync<RoleWithoutAbilitiesResponse>(ruleSetId);
        return Results.Ok(result);
    }
}

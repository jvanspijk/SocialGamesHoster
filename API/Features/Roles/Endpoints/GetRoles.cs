using API.DataAccess.Repositories;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    public static async Task<IResult> HandleAsync(RoleRepository repository)
    {
        List<RoleWithoutAbilitiesResponse> result = await repository.GetAllAsync<RoleWithoutAbilitiesResponse>();
        return Results.Ok(result);
    }
}

using API.DataAccess.Repositories;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    public static async Task<IResult> HandleAsync(RoleRepository repository)
    {
        List<RoleResponse> result = await repository.GetAllAsync<RoleResponse>();
        return Results.Ok(result);
    }
}

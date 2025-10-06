using API.DataAccess.Repositories;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;

public class GetRole
{
    public static async Task<IResult> HandleAsync(RoleRepository repository, int id)
    {
        RoleResponse? result = await repository.GetAsync<RoleResponse>(id);
        if (result == null)
        {
            return Results.NotFound($"Role with id {id} not found.");
        }
        return Results.Ok(result);
    }
}

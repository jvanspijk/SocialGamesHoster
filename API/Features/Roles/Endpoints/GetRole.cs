namespace API.Features.Roles.Endpoints;

public class GetRole
{
    public static async Task<IResult> HandleAsync(RoleService roleService, int id)
    {
        var result = await roleService.GetRoleAsync(id);
        return result.AsIResult();
    }
}

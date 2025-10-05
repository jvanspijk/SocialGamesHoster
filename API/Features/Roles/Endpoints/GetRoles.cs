namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    public static async Task<IResult> HandleAsync(RoleService roleService)
    {
        var result = await roleService.GetAllAsync();
        return result.AsIResult();
    }
}

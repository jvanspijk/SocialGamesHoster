using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public class GetRoles
{
    public record Response(int Id, string Name, string Description)
    : IProjectable<Role, Response>
    {
        public static Expression<Func<Role, Response>> Projection =>
            role => new Response(
                role.Id,
                role.Name,
                role.Description
            );
    }
    public static async Task<IResult> HandleAsync(RoleRepository repository, int ruleSetId)
    {
        List<Response> result = await repository.GetAllFromRulesetAsync<Response>(ruleSetId);
        return Results.Ok(result);
    }
}

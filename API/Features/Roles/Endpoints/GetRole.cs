using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Roles.Common;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public class GetRole
{
    public record Response(int Id, string Name, string Description, List<AbilityInfo> Abilities)
        : IProjectable<Role, Response>
    {
        public static Expression<Func<Role, Response>> Projection =>
            role => new Response(
                role.Id,
                role.Name,
                role.Description,
                role.Abilities.Select(a => new AbilityInfo(a.Id, a.Name, a.Description)).ToList()
            );
    }
    public static async Task<IResult> HandleAsync(RoleRepository repository, int id)
    {
        Response? result = await repository.GetAsync<Response>(id);
        if (result == null)
        {
            return Results.NotFound($"Role with id {id} not found.");
        }
        return Results.Ok(result);
    }
}

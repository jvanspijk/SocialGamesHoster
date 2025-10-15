using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;
using API.Features.Abilities.Responses;

namespace API.Features.Roles.Responses;

public record RoleResponse(int Id, string Name, string Description, List<AbilityResponse> Abilities) 
    : IProjectable<Role, RoleResponse>
{  
    public static Expression<Func<Role, RoleResponse>> Projection =>
        role => new RoleResponse(
            role.Id,
            role.Name,
            role.Description,
            role.Abilities.ProjectTo<Ability, AbilityResponse>().ToList()
        );
}
using API.Features.Abilities.Responses;
using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Roles.Responses;

public readonly record struct RoleResponse(int Id, string Name, string Description, List<AbilityResponse> Abilities) 
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
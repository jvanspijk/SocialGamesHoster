using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Roles.Responses;

public record struct RoleAbilityResponse(int Id, string Name, string Description) : IProjectable<Ability, RoleAbilityResponse>
{
    public static Expression<Func<Ability, RoleAbilityResponse>> Projection => ability => new RoleAbilityResponse(ability.Id, ability.Name, ability.Description);
}

/// <summary>
/// Role with id, name, description and abilities.
/// </summary>
/// <param name="Id"></param>
/// <param name="Name"></param>
/// <param name="Description"></param>
/// <param name="Abilities"></param>
public record struct RoleResponse(int Id, string Name, string Description, List<RoleAbilityResponse> Abilities) : IProjectable<Role, RoleResponse>
{
    public static Expression<Func<Role, RoleResponse>> Projection =>
        role => new RoleResponse(role.Id, role.Name, role.Description, 
            role.Abilities.AsQueryable().Select(RoleAbilityResponse.Projection).ToList());
}


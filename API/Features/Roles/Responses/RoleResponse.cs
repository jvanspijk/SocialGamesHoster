using API.Domain.Models;
using API.Features.Abilities.Responses;

namespace API.Features.Roles.Responses;

public sealed record class RoleResponse(int Id, string Name, string Description)
{
    public RoleResponse(Role roleEntity) : this(roleEntity.Id, roleEntity.Name, roleEntity.Description)
    {     
        Abilities = roleEntity.AbilityAssociations?
            .Select(ra => new AbilityResponse(ra.Ability!))
            .ToList() ?? [];
    }

    public IReadOnlyList<AbilityResponse> Abilities { get; init; } = [];    
}


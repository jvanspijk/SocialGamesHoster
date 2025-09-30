using API.Models;

namespace API.DTO;

public sealed record class RoleDTO(int Id, string Name, string Description)
{
    public RoleDTO(Role roleEntity) : this(roleEntity.Id, roleEntity.Name, roleEntity.Description)
    {       
        VisibleRoleIds = roleEntity.CanSee?
            .Select(rv => rv.VisibleRoleId)
            .ToList() ?? [];

        Abilities = roleEntity.AbilityAssociations?
            .Select(ra => new AbilityDTO(ra.Ability!))
            .ToList() ?? [];
    }

    public IReadOnlyList<int> VisibleRoleIds { get; init; } = [];

    public IReadOnlyList<AbilityDTO> Abilities { get; init; } = [];    
}


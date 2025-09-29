using API.Models;

namespace API.DTO;

public sealed record RoleDTO(int Id, string Name, string Description)
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

    public List<int> VisibleRoleIds { get; set; } = [];

    public List<AbilityDTO> Abilities { get; set; } = [];    
}


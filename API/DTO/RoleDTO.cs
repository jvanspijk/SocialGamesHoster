using API.Models;

namespace API.DTO;

public record RoleDTO
{
    public int Id { get; set; }
    public required string Name { get; set; }
    public string? Description { get; set; }

    public List<int> VisibleRoleIds { get; set; } = [];

    public List<AbilityDTO> Abilities { get; set; } = [];

    public static RoleDTO FromModel(Role roleEntity)
    {
        return new RoleDTO
        {
            Id = roleEntity.Id,
            Name = roleEntity.Name,
            Description = roleEntity.Description,
            VisibleRoleIds = roleEntity.CanSee?
                .Select(rv => rv.VisibleRoleId)
                .ToList() ?? [],

            Abilities = roleEntity.AbilityAssociations?
                .Select(ra => AbilityDTO.FromModel(ra.Ability))
                .ToList() ?? []
        };
    }
}


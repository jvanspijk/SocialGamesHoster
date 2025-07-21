using API.Models;

namespace API.DTO;

public record RoleDTO
{
    public int Id { get; set; }
    public required string Name { get; set; }
    public string? Description { get; set; }

    public List<int> RolesVisibleToRole { get; set; } = [];

    public List<AbilityDTO> Abilities { get; set; } = [];

    public static RoleDTO FromModel(Role roleEntity)
    {
        return new RoleDTO
        {            
            Id = roleEntity.Id,
            Name = roleEntity.Name,
            Description = roleEntity.Description,
            RolesVisibleToRole = roleEntity.RolesVisibleToRole ?? new List<int>(),
            Abilities = roleEntity.AbilityAssociations != null
                ? roleEntity.AbilityAssociations.Select(ra => AbilityDTO.FromModel(ra.Ability)).ToList()
                : new List<AbilityDTO>()
        };
    }
}


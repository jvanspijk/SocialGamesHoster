using API.Models;

namespace API.DTO;

public class RoleDTO
{
    public required string Name { get; set; }
    public string? Description { get; set; }

    public ICollection<AbilityDTO> Abilities { get; set; } = [];

    public static RoleDTO? FromModel(Role? roleEntity)
    {
        if (roleEntity == null)
        {
            return null;
        }

        return new RoleDTO
        {            
            Name = roleEntity.Name,
            Description = roleEntity.Description,
            Abilities = roleEntity.AbilityAssociations != null
                ? roleEntity.AbilityAssociations.Select(ra => AbilityDTO.FromModel(ra.Ability)).ToList()
                : new List<AbilityDTO>()
        };
    }
}


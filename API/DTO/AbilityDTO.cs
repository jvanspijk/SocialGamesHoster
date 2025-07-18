using API.Models;

namespace API.DTO;

public class AbilityDTO
{
    public required string Name { get; set; }
    public required string Description { get; set; }

    public static AbilityDTO FromModel(Ability? ability)
    {
        if (ability == null)
        {
            return null!;
        }

        return new AbilityDTO
        {
            Name = ability.Name,
            Description = ability.Description
        };
    }
}

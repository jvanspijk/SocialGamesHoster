using API.Models;

namespace API.DTO;

public readonly record struct AbilityDTO(string Name, string Description)
{
    public AbilityDTO(Ability ability) : this(ability.Name, ability.Description) { }
}

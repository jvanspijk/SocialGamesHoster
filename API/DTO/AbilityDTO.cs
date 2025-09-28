using API.Models;

namespace API.DTO;

public sealed record AbilityDTO(string Name, string Description)
{
    public AbilityDTO(Ability ability) : this(ability.Name, ability.Description) { }
}

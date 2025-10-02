using API.Models;

namespace API.Features.Abilities.Responses;

public readonly record struct AbilityResponse(string Name, string Description)
{
    public AbilityResponse(Ability ability) : this(ability.Name, ability.Description) { }
}

using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Abilities.Responses;

public readonly record struct AbilityResponse(string Name, string Description) : IProjectable<Ability, AbilityResponse>
{
    public static Expression<Func<Ability, AbilityResponse>> Projection =>
        ability => new AbilityResponse(ability.Name, ability.Description);
}

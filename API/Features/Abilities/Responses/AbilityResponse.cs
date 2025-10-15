using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Abilities.Responses;

public record AbilityResponse(int Id, string Name, string Description) : IProjectable<Ability, AbilityResponse>
{
    public static Expression<Func<Ability, AbilityResponse>> Projection =>
        ability => new AbilityResponse(ability.Id, ability.Name, ability.Description);
}

using API.DataAccess;
using API.Domain.Models;
using API.Features.Abilities.Responses;
using API.Features.Roles.Responses;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Responses;

public record struct RulesetResponse(int Id, string Name, string Description, List<AbilityResponse> Abilities, List<RoleResponse> Roles) : IProjectable<Ruleset, RulesetResponse>
{
    public static Expression<Func<Ruleset, RulesetResponse>> Projection =>
        rs => new RulesetResponse(
                rs.Id, rs.Name, rs.Description,
                rs.Abilities.ProjectTo<Ability, AbilityResponse>().ToList(),
                rs.Roles.ProjectTo<Role, RoleResponse>().ToList()
             );
}

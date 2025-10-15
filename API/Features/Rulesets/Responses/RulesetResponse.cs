using API.DataAccess;
using API.Domain.Models;
using API.Features.Abilities.Responses;
using API.Features.Roles.Responses;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Responses;

public record RulesetRoleResponse(int Id, string Name, string Description, List<int> AbilityIds, List<int> CanSeeRoleIds);

public record RulesetResponse(int Id, string Name, string Description, List<AbilityResponse> Abilities, List<RulesetRoleResponse> Roles) 
    : IProjectable<Ruleset, RulesetResponse>
{
    public static Expression<Func<Ruleset, RulesetResponse>> Projection =>
        rs => new RulesetResponse(
                rs.Id, rs.Name, rs.Description,
                rs.Abilities.ProjectTo<Ability, AbilityResponse>().ToList(),
                rs.Roles.Select(
                    r => new RulesetRoleResponse(
                        r.Id, r.Name, r.Description,
                        r.Abilities.Select(a => a.Id).ToList(),
                        r.KnowsAbout.Where(k => k.KnowledgeType == KnowledgeType.Role).Select(k => k.TargetId).ToList()                        
                    )).ToList()
                );

}

using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Rulesets.Common;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetFullRuleset
{
    public record Response(int Id, string Name, string Description, List<AbilityInfo> Abilities, List<RoleInfo> Roles)
    : IProjectable<Ruleset, Response>
    {
        public static Expression<Func<Ruleset, Response>> Projection =>
            rs => new Response(
                    rs.Id, rs.Name, rs.Description,
                    rs.Abilities.Select(a => new AbilityInfo(a.Id, a.Name, a.Description)).ToList(),
                    rs.Roles.Select(
                        r => new RoleInfo(
                            r.Id, r.Name, r.Description,
                            r.Abilities.Select(a => a.Id).ToList(),
                            r.KnowsAbout.Where(k => k.KnowledgeType == KnowledgeType.Role).Select(k => k.TargetId).ToList()
                        )).ToList()
                    );

    }
    public static async Task<IResult> HandleAsync(RulesetRepository repository, int rulesetId)
    {
        Response? response = await repository.GetAsync<Response>(rulesetId);
        if (response == null)
        {
            return Results.Problem($"Ruleset with id {rulesetId} not found.");
        }
        return Results.Ok(response);
    }
}

using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Abilities.Endpoints;

public static class GetAbilitiesFromRuleset
{
    public record Response(int Id, string Name, string Description) : IProjectable<Ability, Response>
    {
        public static Expression<Func<Ability, Response>> Projection =>
            ability => new Response(ability.Id, ability.Name, ability.Description);
    }
    public static async Task<IResult> HandleAsync(AbilityRepository repository, int rulesetId)
    {
        List<Response> result = await repository.GetAllFromRulesetAsync<Response>(rulesetId);
        return Results.Ok(result);
    }
}

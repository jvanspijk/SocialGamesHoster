using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Rulesets.Endpoints;

public static class GetAllRulesets
{
    public record Response(int Id, string Name, string Description)
        : IProjectable<Ruleset, Response>
    {
        public static Expression<Func<Ruleset, Response>> Projection =>
            rs => new Response(rs.Id, rs.Name, rs.Description);
    }
    public static async Task<IResult> HandleAsync(RulesetRepository repository)
    {
        List<Response> response = await repository.GetAllAsync<Response>();
        return Results.Ok(response);
    }
}

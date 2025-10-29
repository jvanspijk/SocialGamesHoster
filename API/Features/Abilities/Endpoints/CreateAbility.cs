using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Abilities.Endpoints;

public static class CreateAbility
{
    public readonly record struct Request(string Name, string Description);
    public record Response(int Id, string Name, string Description) : IProjectable<Ability, Response>
    {
        public static Expression<Func<Ability, Response>> Projection =>
            ability => new Response(ability.Id, ability.Name, ability.Description);
    }
    public static async Task<IResult> HandleAsync(AbilityRepository repository, Request request, int rulesetId)
    {
        Ability ability = new() { Name = request.Name, Description = request.Description, RulesetId = rulesetId };
        Ability result = await repository.CreateAsync(ability);

        Response response = new(result.Id, result.Name, result.Description);

        return Results.Ok(response);
    }
}

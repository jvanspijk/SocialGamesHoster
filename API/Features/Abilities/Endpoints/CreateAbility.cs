using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
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
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<Ability> repository, Request request, int rulesetId)
    {
        Ability ability = new() { Name = request.Name, Description = request.Description, RulesetId = rulesetId };

        // TODO: validation
        repository.Add(ability);
        await repository.SaveChangesAsync();

        Response response = new(ability.Id, ability.Name, ability.Description);    

        return APIResults.Ok(response);
    }
}

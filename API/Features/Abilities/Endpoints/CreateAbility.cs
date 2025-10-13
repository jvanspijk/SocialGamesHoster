using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Abilities.Responses;

namespace API.Features.Abilities.Endpoints;

public static class CreateAbility
{
    public readonly record struct Request(string Name, string Description, int RulesetId);
    public static async Task<IResult> HandleAsync(AbilityRepository repository, Request request, HttpContext context)
    {
        Ability ability = new() { Name = request.Name, Description = request.Description, RulesetId = request.RulesetId };
        Ability result = await repository.CreateAsync(ability);

        AbilityResponse response = new(result.Id, result.Name, result.Description);
        string uri = context.Request.Scheme + "://" +
            context.Request.Host.ToString() +
            $"/abilities/{result.Id}";

        return Results.Created(uri, response);
    }
}

using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Abilities.Responses;

namespace API.Features.Abilities.Endpoints;

public static class UpdateAbility
{
    public readonly record struct Request(int Id, string? NewName, string? NewDescription);

    public static async Task<IResult> HandleAsync(AbilityRepository repository, Request request)
    {
        Ability? ability = await repository.GetByIdAsync(request.Id);
        if (ability == null)
        {
            return Results.NotFound();
        }
        if (!string.IsNullOrWhiteSpace(request.NewName))
        {
            ability.Name = request.NewName;
        }
        if (!string.IsNullOrWhiteSpace(request.NewDescription))
        {
            ability.Description = request.NewDescription;
        }
        Ability updatedAbility = await repository.UpdateAsync(ability);
        AbilityResponse response = updatedAbility.ProjectTo<Ability, AbilityResponse>().First();
        return Results.Ok(response);
    }
}

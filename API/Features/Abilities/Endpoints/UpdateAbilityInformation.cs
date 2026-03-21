using API.DataAccess;
using API.Domain.Entities;
using System.Linq.Expressions;

namespace API.Features.Abilities.Endpoints;

public static class UpdateAbilityInformation
{
    public readonly record struct Request(string? NewName, string? NewDescription);
    public record Response(int Id, string Name, string Description) : IProjectable<Ability, Response>
    {
        public static Expression<Func<Ability, Response>> Projection =>
            ability => new Response(ability.Id, ability.Name, ability.Description);
    }

    public static async Task<IResult> HandleAsync(IRepository<Ability> repository, int id, Request request)
    {
        Ability? ability = await repository.GetWithTrackingAsync(id);
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

        await repository.SaveChangesAsync();

        var response = ability.ConvertToResponse<Ability, Response>();
        return Results.Ok(response);
    }
}

using API.DataAccess;
using API.Domain.Entities;
using API.Domain.Validation;
using API.Features.Roles.Common;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public static class UpdateRoleAbilities
{
    public readonly record struct Request(List<int> AbilityIds);
    public record Response(int Id, List<AbilityInfo> Abilities)
    : IProjectable<Role, Response>
    {
        public static Expression<Func<Role, Response>> Projection =>
            role => new Response(
                role.Id,
                role.Abilities.Select(a => new AbilityInfo(a.Id, a.Name, a.Description)).ToList()
            );
    }
    public async static Task<IResult> HandleAsync(IRepository<Role> repository, IRepository<Ability> abilityRepository, IMemoryCache cache, int id, Request request)
    {
        Role? role = await repository.GetWithTrackingAsync(id);
        if (role == null)
        {
            return Results.NotFound($"Role with id {id} not found.");
        }

        var abilities = await abilityRepository.GetListWithTrackingAsync(request.AbilityIds);

        var rulesetMismatchErrors = abilities
            .Where(ability => ability.RulesetId != role.RulesetId)
            .Select(ability => new ValidationError(
                "RulesetId",
                $"Ability with ID {ability.Id} must share the role's ruleset. " +
                $"(Ability Ruleset: {ability.RulesetId}, Role Ruleset: {role.RulesetId})"
            ))
            .ToList();

        if (rulesetMismatchErrors.Count != 0)
        {
            return Results.ValidationProblem(rulesetMismatchErrors.ToProblemDetails());
        }

        role.Abilities = abilities;
        await repository.SaveChangesAsync();

        Response response = role.ConvertToResponse<Role, Response>();
        return Results.Ok(response);
    }
}

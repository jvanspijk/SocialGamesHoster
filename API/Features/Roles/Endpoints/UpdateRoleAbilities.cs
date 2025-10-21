using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Domain.Validation;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;

public static class UpdateRoleAbilities
{
    public readonly record struct Request(List<int> AbilityIds);
    public async static Task<IResult> HandleAsync(RoleRepository repository, AbilityRepository abilityRepository, int id, Request request)
    {
        var abilitiesResult = await abilityRepository.GetAsync(request.AbilityIds);
        if (abilitiesResult.IsFailure)
        {
            return abilitiesResult.AsIResult();
        }

        Role? role = await repository.GetAsync(id);
        if (role == null)
        {
            return Results.NotFound($"Role with id {id} not found.");
        }
        var abilities = abilitiesResult.Value;

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
        var updatedRole = await repository.UpdateAsync(role);
        RoleResponse response = updatedRole.ConvertToResponse<Role, RoleResponse>();
        return Results.Ok(response);
    }
}

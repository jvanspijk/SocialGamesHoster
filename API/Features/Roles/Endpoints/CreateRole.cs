using API.DataAccess.Repositories;
using API.Domain.Models;

namespace API.Features.Roles.Endpoints;

public static class CreateRole
{
    public readonly record struct Request(string Name, string? Description);

    public static async Task<IResult> HandleAsync(RoleRepository repository, int rulesetId, Request request)
    {
        // TODO: request validation
        string description = request.Description ?? "";
        Role role = new() { Name = request.Name, Description = description, RulesetId = rulesetId };
        Role createdRole = await repository.CreateAsync(role);
        return Results.Ok(createdRole); // Can this be a problem?
    }
}

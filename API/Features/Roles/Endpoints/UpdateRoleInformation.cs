using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;
public static class UpdateRoleInformation
{
    public readonly record struct Request(string? Name, string? Description);

    public async static Task<IResult> HandleAsync(RoleRepository repository, int id, Request request)
    {
        Role? role = await repository.GetAsync(id);
        if (role == null)
        {
            return Results.NotFound($"Role with id {id} not found.");
        }

        bool hasChanged = false;
        if (request.Name != null)
        {
            role.Name = request.Name;
            hasChanged = true;
        }

        if (request.Description != null)
        {
            role.Description = request.Description;
            hasChanged = true;
        }

        if (!hasChanged)
        {
            return Results.NoContent();
        }

        var updatedRole = await repository.UpdateAsync(role);
        RoleResponse response = updatedRole.ConvertToResponse<Role, RoleResponse>();
        return Results.Ok(response);
    }
}

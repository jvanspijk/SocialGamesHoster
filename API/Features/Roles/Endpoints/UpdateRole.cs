using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Roles.Responses;

namespace API.Features.Roles.Endpoints;
public static class UpdateRole
{
    public readonly record struct Request(int Id, string? Name, string? Description);

    public async static Task<IResult> HandleAsync(RoleRepository repository, Request request)
    {
        Role? role = await repository.GetAsync(request.Id);
        if (role == null)
        {
            return Results.NotFound($"Role with id {request.Id} not found.");
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

        await repository.UpdateAsync(role);
        RoleResponse response = role.ProjectTo<Role, RoleResponse>().First();
        return Results.Ok(response);
    }
}

using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;
public static class UpdateRoleInformation
{
    public readonly record struct Request(string? Name, string? Description);

    public record Response(int Id, string Name, string Description)
    : IProjectable<Role, Response>
    {
        public static Expression<Func<Role, Response>> Projection =>
            role => new Response(
                role.Id,
                role.Name,
                role.Description               
            );
    }

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
        Response response = updatedRole.ConvertToResponse<Role, Response>();
        return Results.Ok(response);
    }
}

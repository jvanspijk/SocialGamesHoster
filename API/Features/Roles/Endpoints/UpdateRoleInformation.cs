using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
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

    public async static Task<Results<Ok<Response>, NoContent, ProblemHttpResult>> HandleAsync(IRepository<Role> repository, IMemoryCache cache, int id, Request request)
    {
        Role? role = await repository.GetWithTrackingAsync(id);
        if (role == null)
        {
            return APIResults.NotFound<Role>(id);
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
            return APIResults.NoContent();
        }

        await repository.SaveChangesAsync();

        Response response = role.ConvertToResponse<Role, Response>();
        return APIResults.Ok(response);
    }
}

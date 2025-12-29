using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public static class CreateRole
{
    public readonly record struct Request(string Name, string? Description) : IValidatable<Request>
    {
        public readonly IEnumerable<ValidationError> Validate()
        {            
            if (string.IsNullOrWhiteSpace(Name) || Name.Length is < 1 or > 32)
            {
                 yield return new ValidationError(nameof(Name), "Role name must be between 1 and 32 characters long.");
            }
            if (!string.IsNullOrWhiteSpace(Description) && Description.Length is > 256)
            {
                yield return new ValidationError(nameof(Description), "Role description must be less than 256 characters long.");
            }
        }
    }

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
    public static async Task<IResult> HandleAsync(RoleRepository repository, IMemoryCache cache, int rulesetId, Request request)
    {
        var validationResult = request.Validate();
        if (validationResult.HasErrors())
        {
            return Results.ValidationProblem(validationResult.ToProblemDetails());
        }
        string description = request.Description ?? "";
        Role role = new() { Name = request.Name, Description = description, RulesetId = rulesetId };
        Role createdRole = await repository.CreateAsync(role); // Should this be a result type?

        GetRoles.InvalidateCache(cache, rulesetId);

        Response response = new(createdRole.Id, createdRole.Name, createdRole.Description);
        return Results.Ok(response); 
    }
}

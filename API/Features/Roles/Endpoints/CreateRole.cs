using API.DataAccess;
using API.Domain.Entities;
using System.ComponentModel.DataAnnotations;
using System.Linq.Expressions;

namespace API.Features.Roles.Endpoints;

public static class CreateRole
{
    public readonly record struct Request(string Name, string? Description) : IValidatableObject
    {
        public IEnumerable<ValidationResult> Validate(ValidationContext validationContext)
        {            
            if (string.IsNullOrWhiteSpace(Name) || Name.Length is < 1 or > 32)
            {
                 yield return new ValidationResult("Role name must be between 1 and 32 characters long.", [nameof(Name)]);
            }
            if (!string.IsNullOrWhiteSpace(Description) && Description.Length is > 256)
            {
                yield return new ValidationResult("Role description must be less than 256 characters long.", [nameof(Description)]);
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
    public static async Task<IResult> HandleAsync(IRepository<Role> repository, int rulesetId, Request request)
    {
        string description = request.Description ?? "";
        Role role = new() { Name = request.Name, Description = description, RulesetId = rulesetId };
        repository.Add(role);
        await repository.SaveChangesAsync();
        Response response = new(role.Id, role.Name, role.Description);
        return Results.Ok(response); 
    }
}

using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Roles.Responses;

public readonly record struct RoleWithoutAbilitiesResponse(int Id, string Name, string Description)
    : IProjectable<Role, RoleWithoutAbilitiesResponse>
{
    public static Expression<Func<Role, RoleWithoutAbilitiesResponse>> Projection =>
        role => new RoleWithoutAbilitiesResponse(
            role.Id,
            role.Name,
            role.Description
        );
}


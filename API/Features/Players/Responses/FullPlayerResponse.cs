using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;

public record struct AbilityResponse(int Id, string Name, string Description) : IProjectable<Ability, AbilityResponse>
{
    public static Expression<Func<Ability, AbilityResponse>> Projection =>
        ability => new AbilityResponse(ability.Id, ability.Name, ability.Description);
}
public record struct RoleWithoutAbilities(int Id, string Name, string Description) : IProjectable<Role, RoleWithoutAbilities>
{
    public static Expression<Func<Role, RoleWithoutAbilities>> Projection =>
    role => new RoleWithoutAbilities(
        role.Id,
        role.Name,
        role.Description
    );
}
public record struct RoleWithAbilities(int Id, string Name, string Description, List<AbilityResponse> Abilities) : IProjectable<Role, RoleWithAbilities>
{
    public static Expression<Func<Role, RoleWithAbilities>> Projection =>
        role => new RoleWithAbilities(
            role.Id,
            role.Name,
            role.Description,
            role.Abilities
                .AsQueryable()
                .Select(AbilityResponse.Projection)
                .ToList()
        );
}
public record struct PartialPlayerResponse(int Id, string Name, RoleWithoutAbilities? Role) : IProjectable<Player, PartialPlayerResponse>
{
    public static Expression<Func<Player, PartialPlayerResponse>> Projection => 
        player => new PartialPlayerResponse(player.Id, player.Name, new RoleWithoutAbilities(player.Role.Id, player.Role.Name, player.Role.Description));
}
public record struct FullPlayerResponse(int Id, string Name, RoleWithAbilities Role, List<RoleWithoutAbilities> CanSee);


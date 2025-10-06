using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.AdminFeatures.Players.Responses;

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
        player => new PartialPlayerResponse(
        player.Id,
        player.Name,
        new RoleWithoutAbilities(player.Role.Id, player.Role.Name, player.Role.Description)
    );
}
public record struct FullPlayerResponse(int Id, string Name, RoleWithAbilities Role, List<PartialPlayerResponse> CanSee, List<PartialPlayerResponse> CanBeSeenBy) : IProjectable<Player, FullPlayerResponse>
{
    public static Expression<Func<Player, FullPlayerResponse>> Projection =>
        player => new FullPlayerResponse(
            player.Id, player.Name,
            new RoleWithAbilities(
                player.Role.Id, player.Role.Name, player.Role.Description,
                player.Role.Abilities
                .AsQueryable()
                .Select(AbilityResponse.Projection)
                .ToList()),
            player.CanSee
            .AsQueryable()
            .Select(PartialPlayerResponse.Projection)
            .ToList(),
            player.CanBeSeenBy
            .AsQueryable()
            .Select(PartialPlayerResponse.Projection)
            .ToList()
        );
}


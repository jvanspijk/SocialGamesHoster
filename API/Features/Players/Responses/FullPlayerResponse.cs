using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;
public record struct PlayerAbilityResponse(int Id, string Name, string Description) : IProjectable<Ability, PlayerAbilityResponse>
{
    public static Expression<Func<Ability, PlayerAbilityResponse>> Projection =>
        ability => new PlayerAbilityResponse(ability.Id, ability.Name, ability.Description);
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
public record struct RoleWithAbilities(int Id, string Name, string Description, List<PlayerAbilityResponse> Abilities) : IProjectable<Role, RoleWithAbilities>
{
    public static Expression<Func<Role, RoleWithAbilities>> Projection =>
        role => new RoleWithAbilities(
            role.Id,
            role.Name,
            role.Description,
            role.Abilities
                .AsQueryable()
                .Select(PlayerAbilityResponse.Projection)
                .ToList()
        );
}
public record struct PartialPlayerResponse(int Id, string Name, RoleWithoutAbilities? Role) : IProjectable<Player, PartialPlayerResponse>
{
    public static Expression<Func<Player, PartialPlayerResponse>> Projection => 
        player => new PartialPlayerResponse(player.Id, player.Name, new RoleWithoutAbilities(player.Role.Id, player.Role.Name, player.Role.Description));
}
public record struct FullPlayerResponse(int Id, string Name, RoleWithAbilities Role, List<RoleWithoutAbilities> CanSee);


using API.DataAccess;
using API.Domain.Models;
using API.Features.Abilities.Responses;
using API.Features.Roles.Responses;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;

public record PlayerResponse(int Id, string Name, RoleResponse? Role) : IProjectable<Player, PlayerResponse>
{
    public static Expression<Func<Player, PlayerResponse>> Projection =>
        player  => new PlayerResponse(
            player.Id,
            player.Name,
            player.Role == null ? null :
            new RoleResponse(player.Role.Id, player.Role.Name, player.Role.Description, 
                player.Role.Abilities.ProjectTo<Ability, AbilityResponse>().ToList()));
}


using API.DataAccess;
using API.Domain.Models;
using API.Features.Roles.Responses;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;

public readonly record struct PlayerResponse(int Id, string Name, RoleResponse? Role) : IProjectable<Player, PlayerResponse>
{
    public static Expression<Func<Player, PlayerResponse>> Projection =>
        player => new PlayerResponse(
            player.Id,
            player.Name,
            player.Role == null ? null :
            player.Role.ProjectTo<Role, RoleResponse>().First()
        );
}


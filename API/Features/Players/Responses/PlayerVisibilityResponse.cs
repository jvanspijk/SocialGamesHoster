using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;

public readonly record struct PlayerVisibilityResponse(List<PlayerNameResponse> CanSeeIds, 
    List<PlayerNameResponse> CanBeSeenByIds) : IProjectable<Player, PlayerVisibilityResponse>
{
    public static Expression<Func<Player, PlayerVisibilityResponse>> Projection =>
        player => new PlayerVisibilityResponse(
            player.CanSee.ProjectTo<Player, PlayerNameResponse>().ToList(),
            player.CanBeSeenBy.ProjectTo<Player, PlayerNameResponse>().ToList()                
        );
}


using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;

public readonly record struct PlayerNameResponse(int Id, string Name) : IProjectable<Player, PlayerNameResponse>
{
    public static Expression<Func<Player, PlayerNameResponse>> Projection =>
        player => new PlayerNameResponse(player.Id, player.Name);
}


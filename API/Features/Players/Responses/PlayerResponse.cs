using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Players.Responses;

public readonly record struct PlayerResponse(int Id, string Name) : IProjectable<Player, PlayerResponse>
{
    public static Expression<Func<Player, PlayerResponse>> Projection => p => new PlayerResponse(p.Id, p.Name);
}

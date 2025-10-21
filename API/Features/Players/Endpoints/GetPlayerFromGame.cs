using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Players.Common;
using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;

public static class GetPlayerFromGame
{
    public record Response(int Id, string Name, RoleInfo? Role) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(
                player.Id,
                player.Name,
                player.Role == null ? null :
                new RoleInfo(player.Role.Id, player.Role.Name, player.Role.Description,
                    player.Role.Abilities
                    .Select(a => new AbilityInfo(a.Id, a.Name, a.Description))
                    .ToList()
                )
            );
    }
    public static async Task<IResult> HandleAsync(PlayerRepository repository, string name, int gameId)
    {
        Response? result = await repository.GetByNameAsync<Response>(name, gameId);
        if (result == null)
        {
            return Results.NotFound($"Player with name '{name}' not found.");
        }
        return Results.Ok(result);
    }
}

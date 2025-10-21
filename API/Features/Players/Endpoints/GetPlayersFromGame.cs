using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;

using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;

public static class GetPlayersFromGame
{
    public record Response(int Id, string Name) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(player.Id, player.Name);
    }
    public static async Task<IResult> HandleAsync(PlayerRepository repository, int gameId)
    {
        List<Response> result = await repository.GetAllFromGameAsync<Response>(gameId);
        return Results.Ok(result);
    }
}

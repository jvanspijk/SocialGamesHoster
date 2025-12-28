using API.DataAccess.Repositories;
using API.Domain;

namespace API.Features.Players.Endpoints;

public static class PlayerExistsInGame
{
    public record Response(bool Exists);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, int playerId, int gameId)
    {
        var player = await repository.GetAsync(playerId);

        if(player == null)
        {
            return Results.Ok(new Response(false));
        }
        if(player.GameId != gameId)
        {
            return Results.Ok(new Response(false));
        }

        return Results.Ok(new Response(true));
    }
}

using API.DataAccess.Repositories;
using API.Domain.Models;

namespace API.Features.Players.Endpoints;
public class DeletePlayer
{
    public static async Task<IResult> HandleAsync(PlayerRepository repository, int id)
    {
        // TODO: validation of request
        Player? playerToDelete = await repository.GetAsync(id);
        if (playerToDelete == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
            
        await repository.DeleteAsync(playerToDelete);
        return Results.NoContent();
    }
}

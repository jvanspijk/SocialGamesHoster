using API.DataAccess;
using API.Domain.Entities;

namespace API.Features.Players.Endpoints;
public static class DeletePlayer
{
    public static async Task<IResult> HandleAsync(IRepository<Player> repository, int id)
    {
        // TODO: validation of request
        Player? playerToDelete = await repository.GetWithTrackingAsync(id);
        if (playerToDelete == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
            
        repository.Remove(playerToDelete);
        await repository.SaveChangesAsync();

        return Results.NoContent();
    }
}

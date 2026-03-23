using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Players.Endpoints;
public static class DeletePlayer
{
    public static async Task<Results<NoContent, ProblemHttpResult>> HandleAsync(IRepository<Player> repository, int id)
    {
        // TODO: validation of request
        Player? playerToDelete = await repository.GetWithTrackingAsync(id);
        if (playerToDelete == null)
        {
            return APIResults.NotFound<Player>(id);
        }
            
        repository.Remove(playerToDelete);
        await repository.SaveChangesAsync();

        return APIResults.NoContent();
    }
}

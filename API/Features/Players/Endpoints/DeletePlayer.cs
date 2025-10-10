using API.DataAccess.Repositories;
using API.Domain.Models;

namespace API.Features.Players.Endpoints;
public class DeletePlayer
{
    public readonly record struct Request(int Id);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request)
    {
        // TODO: validation of request
        Player? playerToDelete = await repository.GetAsync(request.Id);
        if (playerToDelete == null)
        {
            return Results.NotFound($"Player with id {request.Id} not found.");
        }
            
        await repository.DeleteAsync(playerToDelete);
        return Results.NoContent();
    }
}

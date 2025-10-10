using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Players.Requests;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class UpdatePlayer
{
    public readonly record struct Request(int Id, string? NewName, int? NewRoleId);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request)
    {
        // TODO: validation

        Player? player = await repository.GetAsync(request.Id);

        if (player == null)
        {
            return Results.NotFound($"Player with ID {request.Id} not found.");
        }

        bool playerChanged = false;

        if(!string.IsNullOrWhiteSpace(request.NewName) && request.NewName != player.Name)
        {
            player.Name = request.NewName;
            playerChanged = true;
        }

        if(request.NewRoleId.HasValue && request.NewRoleId != player.RoleId)
        {
            player.RoleId = request.NewRoleId;
            playerChanged = true;
        }

        if(!playerChanged)
        {
            return Results.NoContent();
        }

        await repository.UpdateAsync(player);
        PlayerResponse response = player.ProjectTo<Player, PlayerResponse>().First();

        return Results.Ok(response);
    }
}

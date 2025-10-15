using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;

public static class UpdatePlayer
{
    public readonly record struct Request(string? NewName, int? NewRoleId);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, int id, Request request)
    {
        // TODO: validation

        Player? player = await repository.GetAsync(id);

        if (player == null)
        {
            return Results.NotFound($"Player with ID {id} not found.");
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

        var updatedPlayer = await repository.UpdateAsync(player);
        PlayerResponse response = updatedPlayer.ConvertToResponse<Player, PlayerResponse>();

        return Results.Ok(response);
    }
}

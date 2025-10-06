using API.DataAccess.Repositories;
using API.Domain.Models;

namespace API.AdminFeatures.Players.Endpoints;

public static class UpdatePlayer
{
    public readonly record struct PlayerResponse(int Id, string Name, int? RoleId);
    public readonly record struct Request(int Id, string? NewName, int? NewRoleId);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request)
    {
        // TODO: validation

        Player? player = await repository.GetAsync(request.Id);

        if (player == null)
        {
            return Results.NotFound($"Player with ID {request.Id} not found.");
        }

        if(request.NewRoleId is not null)
        {
            player.RoleId = request.NewRoleId;
        }

        if (request.NewName is not null)
        {
            player.Name = request.NewName;
        }

        await repository.UpdateAsync(player);

        return Results.Ok(new PlayerResponse(player.Id, player.Name, player.RoleId));
    }
}

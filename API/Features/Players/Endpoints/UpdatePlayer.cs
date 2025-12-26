using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Players.Common;
using API.Features.Players.Hubs;
using Microsoft.AspNetCore.SignalR;
using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;

public static class UpdatePlayer
{
    public readonly record struct Request(string? NewName, int? NewRoleId);
    public record Response(int Id, string Name, int? RoleId) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(
                player.Id,
                player.Name,
                player.RoleId
            );
    }
    public static async Task<IResult> HandleAsync(PlayerRepository repository, IHubContext<PlayerHub, IPlayerHub> hub, int id, Request request)
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
            return Results.Ok(player.ConvertToResponse<Player, Response>());
        }

        var updatedPlayer = await repository.UpdateAsync(player);
        await PlayerHub.NotifyPlayerUpdated(hub, updatedPlayer.Id);

        Response response = new(updatedPlayer.Id, updatedPlayer.Name, updatedPlayer.RoleId);
        return Results.Ok(response);
    }
}

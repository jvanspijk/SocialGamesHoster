using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Players.Common;
using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;

public static class UpdatePlayer
{
    public readonly record struct Request(string? NewName, int? NewRoleId);
    public record Response(int Id, string Name, RoleInfo? Role) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(
                player.Id,
                player.Name,
                player.Role == null ? null :
                new RoleInfo(player.Role.Id, player.Role.Name, player.Role.Description,
                    player.Role.Abilities
                    .Select(a => new AbilityInfo(a.Id, a.Name, a.Description))
                    .ToList()
                )
            );
    }
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
            return Results.Ok(player.ConvertToResponse<Player, Response>());
        }

        var updatedPlayer = await repository.UpdateAsync(player);
        Response response = updatedPlayer.ConvertToResponse<Player, Response>();

        return Results.Ok(response);
    }
}

using API.Domain.Models;

namespace API.Features.Players.Responses;

/// <summary>
/// Player DTO without their role information.
/// </summary>
/// <param name="Id"></param>
/// <param name="Name"></param>
public readonly record struct PlayerWithoutRoleResponse(int Id, string Name)
{
    public PlayerWithoutRoleResponse(Player player) : this(player.Id, player.Name) { }
}

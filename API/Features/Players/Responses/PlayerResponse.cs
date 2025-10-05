using API.Domain.Models;

namespace API.Features.Players.Responses;

/// <summary>
/// Player DTO without their role information.
/// </summary>
/// <param name="Id"></param>
/// <param name="Name"></param>
public readonly record struct PlayerResponse(int Id, string Name)
{
    public PlayerResponse(Player player) : this(player.Id, player.Name) { }
}

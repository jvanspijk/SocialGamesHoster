using API.Features.Roles.Responses;
using API.Models;

namespace API.Features.Players.Responses;
/// <summary>
/// Player DTO with their role information.
/// </summary>
/// <param name="Id"></param>
/// <param name="Name"></param>
public record PlayerWithRoleResponse(int Id, string Name)
{
    public required RoleResponse Role { get; init; }
    public PlayerWithRoleResponse(Player player) : this(player.Id, player.Name)
    {       
        if (player.Role is null)
        {
            throw new InvalidOperationException($"Cannot create PlayerDetailDTO: Role is required but null for Player ID {player.Id}");
        }

        Role = new RoleResponse(player.Role);
    }
}

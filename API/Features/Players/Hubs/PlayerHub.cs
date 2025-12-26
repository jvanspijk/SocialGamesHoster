using Microsoft.AspNetCore.SignalR;

namespace API.Features.Players.Hubs;

public interface IPlayerHub
{
    Task PlayerUpdated(int playerId);
}

public class PlayerHub : Hub<IPlayerHub>
{
    public static async Task NotifyPlayerUpdated(IHubContext<PlayerHub, IPlayerHub> hubContext, int playerId)
    {
        await hubContext.Clients.All.PlayerUpdated(playerId);
    }
}

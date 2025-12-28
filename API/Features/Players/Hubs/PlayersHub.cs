using Microsoft.AspNetCore.SignalR;

namespace API.Features.Players.Hubs;

public interface IPlayersHub
{
    Task PlayerUpdated(int playerId);
}

public class PlayersHub : Hub<IPlayersHub>
{
    public static async Task NotifyPlayerUpdated(IHubContext<PlayersHub, IPlayersHub> hubContext, int playerId)
    {
        await hubContext.Clients.All.PlayerUpdated(playerId);
    }
}

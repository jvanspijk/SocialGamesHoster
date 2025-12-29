using Microsoft.AspNetCore.SignalR;

namespace API.Features.GameSessions.Hubs;

public interface IGameSessionsHub
{
    Task GameSessionUpdated(int gameSessionId);
}

public class GameSessionsHub : Hub<IGameSessionsHub>
{
    public static async Task NotifyGameSessionUpdated(IHubContext<GameSessionsHub, IGameSessionsHub> hubContext, int gameSessionId)
    {
        await hubContext.Clients.All.GameSessionUpdated(gameSessionId);
    }
}
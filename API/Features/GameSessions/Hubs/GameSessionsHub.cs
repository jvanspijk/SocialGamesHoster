using Microsoft.AspNetCore.SignalR;

namespace API.Features.GameSessions.Hubs;

public interface IGameSessionsHub
{
    Task OnRoundStarted(int gameSessionId);
    Task OnRoundEnded(int gameSessionId);
    Task GameSessionUpdated(int gameSessionId);
}

public class GameSessionsHub : Hub<IGameSessionsHub>
{
    public static async Task NotifyGameSessionUpdated(IHubContext<GameSessionsHub, IGameSessionsHub> hubContext, int gameSessionId)
    {
        await hubContext.Clients.All.GameSessionUpdated(gameSessionId);
    }
    public static async Task NotifyRoundStarted(IHubContext<GameSessionsHub, IGameSessionsHub> hubContext, int gameSessionId)
    {
        await hubContext.Clients.All.OnRoundStarted(gameSessionId);
    }
    public static async Task NotifyRoundEnded(IHubContext<GameSessionsHub, IGameSessionsHub> hubContext, int gameSessionId)
    {
        await hubContext.Clients.All.OnRoundEnded(gameSessionId);
    }
}
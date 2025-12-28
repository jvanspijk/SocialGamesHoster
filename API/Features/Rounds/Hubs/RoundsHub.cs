using Microsoft.AspNetCore.SignalR;

namespace API.Features.Rounds.Hubs;

public interface IRoundsHub
{
    Task OnRoundStarted(int roundId);
    Task OnRoundEnded(int roundId);
}
public class RoundsHub : Hub<IRoundsHub>
{
    public static async Task NotifyRoundStarted(IHubContext<RoundsHub, IRoundsHub> hubContext, int roundId)
    {
        await hubContext.Clients.All.OnRoundStarted(roundId);
    }
    public static async Task NotifyRoundEnded(IHubContext<RoundsHub, IRoundsHub> hubContext, int roundId)
    {
        await hubContext.Clients.All.OnRoundEnded(roundId);
    }
}

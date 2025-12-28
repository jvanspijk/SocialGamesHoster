using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Hubs;

public interface ITimersHub
{
    Task OnTimerStarted(int remainingSeconds, int totalSeconds);
    Task OnTimerChanged(int remainingSeconds, int totalSeconds);
    Task OnTimerPaused(int remainingSeconds, int totalSeconds);
    Task OnTimerResumed(int remainingSeconds, int totalSeconds);
    Task OnTimerStopped();
    Task OnTimerFinished();
}
public class TimersHub : Hub<ITimersHub>
{
    public static async Task NotifyTimerChanged(IHubContext<TimersHub, ITimersHub> hubContext, int remainingSeconds, int totalSeconds)
    {
        await hubContext.Clients.All.OnTimerChanged(remainingSeconds, totalSeconds);
    }
    public static async Task NotifyTimerPaused(IHubContext<TimersHub, ITimersHub> hubContext, int remainingSeconds, int totalSeconds)
    {
        await hubContext.Clients.All.OnTimerPaused(remainingSeconds, totalSeconds);
    }
    public static async Task NotifyTimerResumed(IHubContext<TimersHub, ITimersHub> hubContext, int remainingSeconds, int totalSeconds)
    {
        await hubContext.Clients.All.OnTimerResumed(remainingSeconds, totalSeconds);
    }
    public static async Task NotifyTimerStopped(IHubContext<TimersHub, ITimersHub> hubContext)
    {
        await hubContext.Clients.All.OnTimerStopped();
    }
    public static async Task NotifyTimerFinished(IHubContext<TimersHub, ITimersHub> hubContext)
    {
        await hubContext.Clients.All.OnTimerFinished();
    }
    public static async Task NotifyTimerStarted(IHubContext<TimersHub, ITimersHub> hubContext, int remainingSeconds, int totalSeconds)
    {
        await hubContext.Clients.All.OnTimerStarted(remainingSeconds, totalSeconds);
    }
}

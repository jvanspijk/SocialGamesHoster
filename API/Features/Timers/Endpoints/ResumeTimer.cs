using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;

public static class ResumeTimer
{
    public record Response(int RemainingSeconds);
    public static async Task<IResult> HandleAsync(RoundTimer timer, IHubContext<TimersHub, ITimersHub> hub)
    {
        timer.Resume();
        await TimersHub.NotifyTimerResumed(hub, (int)timer.RemainingTime.TotalSeconds, (int)timer.TotalTime.TotalSeconds);
        return Results.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
    }
}

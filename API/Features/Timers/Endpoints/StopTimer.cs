using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;

public static class StopTimer
{    
    public static async Task<IResult> HandleAsync(RoundTimer timer, IHubContext<TimersHub, ITimersHub> hub)
    {
        timer.Stop();
        await TimersHub.NotifyTimerStopped(hub);
        return Results.NoContent();
    }
}
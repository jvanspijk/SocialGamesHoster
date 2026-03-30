using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;

public static class StopTimer
{    
    public static async Task<Results<NoContent, ProblemHttpResult>> HandleAsync(IGameTimer timer, IHubContext<TimersHub, ITimersHub> hub)
    {
        timer.Stop();
        await TimersHub.NotifyTimerStopped(hub);
        return APIResults.NoContent();
    }
}
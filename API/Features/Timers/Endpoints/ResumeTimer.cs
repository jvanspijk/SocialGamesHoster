using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;

public static class ResumeTimer
{
    public record Response(int RemainingSeconds);
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IGameTimer timer, IHubContext<TimersHub, ITimersHub> hub)
    {
        timer.Resume();
        await TimersHub.NotifyTimerResumed(hub, (int)timer.RemainingTime.TotalSeconds, (int)timer.TotalTime.TotalSeconds);
        return APIResults.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
    }
}

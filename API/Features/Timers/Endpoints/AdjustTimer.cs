using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;
public static class AdjustTimer
{
    public readonly record struct Request(int DeltaSeconds);
    public readonly record struct Response(int RemainingSeconds);
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IGameTimer timer, IHubContext<TimersHub, ITimersHub> hub, Request request)
    {
        if(timer.RemainingTime.TotalSeconds + request.DeltaSeconds < 0)
        {                
            return APIResults.BadRequest("Cannot reduce time below zero.");                
        }

        timer.Adjust(TimeSpan.FromSeconds(request.DeltaSeconds));
        await TimersHub.NotifyTimerChanged(hub, (int)timer.RemainingTime.TotalSeconds, (int)timer.TotalTime.TotalSeconds);
        return APIResults.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
    }
}

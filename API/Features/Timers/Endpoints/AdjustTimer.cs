using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;
public static class AdjustTimer
{
    public readonly record struct Request(int DeltaSeconds);
    public readonly record struct Response(int RemainingSeconds);
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(RoundTimer timer, IHubContext<TimersHub, ITimersHub> hub, Request request)
    {
        //TODO: fix the logic below
        if(timer.RemainingTime.TotalSeconds - request.DeltaSeconds < 0)
        {                
            return APIResults.BadRequest("Cannot reduce time below zero.");                
        }

        if (timer.State == TimerState.Stopped)
        {
            timer.Start(TimeSpan.FromSeconds(request.DeltaSeconds));
            return APIResults.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
        }

        timer.AdjustTime(TimeSpan.FromSeconds(request.DeltaSeconds));
        await TimersHub.NotifyTimerChanged(hub, (int)timer.RemainingTime.TotalSeconds, (int)(timer.RemainingTime.TotalSeconds + request.DeltaSeconds));
        return APIResults.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
    }
}

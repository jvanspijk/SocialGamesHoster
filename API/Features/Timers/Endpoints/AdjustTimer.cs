using API.Domain;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;
public static class AdjustTimer
{
    public record Request(int DeltaSeconds);
    public record Response(int RemainingSeconds);
    public static async Task<IResult> HandleAsync(RoundTimer timer, IHubContext<TimersHub, ITimersHub> hub, Request request)
    {
        //TODO: fix the logic below
        if(timer.RemainingTime.TotalSeconds - request.DeltaSeconds < 0)
        {                
            return Results.BadRequest("Cannot reduce time below zero.");                
        }

        if (timer.State == TimerState.Stopped)
        {
            timer.Start(TimeSpan.FromSeconds(request.DeltaSeconds));
            return Results.Ok(timer.RemainingTime.TotalSeconds);
        }

        timer.AdjustTime(TimeSpan.FromSeconds(request.DeltaSeconds));
        await TimersHub.NotifyTimerChanged(hub, (int)timer.RemainingTime.TotalSeconds, (int)(timer.RemainingTime.TotalSeconds + request.DeltaSeconds));
        return Results.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
    }
}

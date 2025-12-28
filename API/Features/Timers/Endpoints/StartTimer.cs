using API.Domain;
using API.Domain.Models;
using API.Features.Timers.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers.Endpoints;

public static class StartTimer
{
    public record Request(int DurationSeconds);
    public record Response(int RemainingSeconds);
    public static async Task<IResult> HandleAsync(RoundTimer timer, IHubContext<TimersHub, ITimersHub> hub, Request request)
    {
        timer.Start(TimeSpan.FromSeconds(request.DurationSeconds));
        await TimersHub.NotifyTimerStarted(hub, (int)timer.RemainingTime.TotalSeconds, (int)timer.TotalTime.TotalSeconds);
        return Results.Ok(new Response((int)timer.RemainingTime.TotalSeconds));
    }
}

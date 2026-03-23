using API.Domain;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Timers.Endpoints;

public class GetTimerState
{
    public readonly record struct Response(int RemainingSeconds, int TotalSeconds, bool IsRunning);
    public static Ok<Response> Handle(RoundTimer timer)
    {
        return APIResults.Ok(new Response(
            (int)timer.RemainingTime.TotalSeconds,
            (int)timer.TotalTime.TotalSeconds,
            timer.State == TimerState.Running));
    }
}

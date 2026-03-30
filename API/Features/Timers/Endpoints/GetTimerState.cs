using API.Domain;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Timers.Endpoints;

public class GetTimerState
{
    public readonly record struct Response(int RemainingSeconds, int TotalSeconds, bool IsRunning);
    public static Results<Ok<Response>, ProblemHttpResult> Handle(IGameTimer timer)
    {
        if (timer.CurrentState == TimerState.Inactive)
        {
            return APIResults.NotFound("No active timer exists.");
        }

        return APIResults.Ok(new Response(
            (int)timer.RemainingTime.TotalSeconds,
            (int)timer.TotalTime.TotalSeconds,
            timer.CurrentState == TimerState.Running));
    }
}

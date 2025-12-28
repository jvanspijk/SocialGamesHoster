using API.Domain;

namespace API.Features.Timers.Endpoints;

public class GetTimerState
{
    public record Response(int RemainingSeconds, int TotalSeconds, bool IsRunning);
    public static IResult Handle(RoundTimer timer)
    {
        return Results.Ok(new Response(
            (int)timer.RemainingTime.TotalSeconds,
            (int)timer.TotalTime.TotalSeconds,
            timer.State == TimerState.Running));
    }
}

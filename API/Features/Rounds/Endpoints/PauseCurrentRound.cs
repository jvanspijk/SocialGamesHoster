using API.Domain;

namespace API.Features.Rounds.Endpoints;

public static class PauseCurrentRound
{
    public static IResult Handle(RoundTimer timer)
    {
        if (timer.State != TimerState.Running)
        {
            return Results.BadRequest("No active round to pause.");
        }

        timer.Pause();
        return Results.Ok(timer.RemainingTime.TotalSeconds);
    }
}

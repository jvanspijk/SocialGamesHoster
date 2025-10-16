using API.Domain;

namespace API.Features.Rounds.Endpoints;

public static class ResumeCurrentRound
{
    public static IResult Handle(RoundTimer timer)
    {
        if (timer.State == TimerState.Running)
        {
            return Results.BadRequest("Round is already running.");
        }

        timer.Resume();
        return Results.Ok(timer.RemainingTime.TotalSeconds);
    }
}

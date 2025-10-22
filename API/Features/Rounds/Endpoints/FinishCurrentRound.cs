using API.Domain;

namespace API.Features.Rounds.Endpoints;

public static class FinishCurrentRound
{
    public static IResult Handle(RoundTimer timer)
    {
        if (timer.State != TimerState.Running)
        {
            return Results.BadRequest("No active round to finish.");
        }

        timer.Finish(); // Let the timer update the database through the notify event

        return Results.NoContent();
    }
}

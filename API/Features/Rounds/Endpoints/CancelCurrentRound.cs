using API.DataAccess.Repositories;
using API.Domain;

namespace API.Features.Rounds.Endpoints;

public static class CancelCurrentRound
{    
    public static IResult Handle(RoundTimer timer)
    {
        if (timer.State != TimerState.Running)
        {
            return Results.BadRequest("No active round to cancel.");
        }       

        timer.Cancel();

        return Results.NoContent();
    }
}
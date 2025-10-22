using API.DataAccess.Repositories;
using API.Domain;

namespace API.Features.Rounds.Endpoints;
public static class AdjustRoundTime
{
    public static async Task<IResult> HandleAsync(RoundRepository repository, RoundTimer timer, int gameId, int deltaSeconds)
    {
        var roundResult = await repository.GetCurrentRoundFromGame(gameId);
        if (roundResult.IsFailure)
        {
            return roundResult.AsIResult();
        }

        if(timer.RemainingTime.TotalSeconds - deltaSeconds < 0)
        {                
            return Results.BadRequest("Cannot reduce time below zero.");                
        }

        if (timer.State == TimerState.Stopped)
        {
            timer.Start(TimeSpan.FromSeconds(deltaSeconds), roundResult.Value.GameId);
            return Results.Ok(timer.RemainingTime.TotalSeconds);
        }

        timer.AdjustTime(TimeSpan.FromSeconds(deltaSeconds));

        return Results.Ok(timer.RemainingTime.TotalSeconds);
    }
}

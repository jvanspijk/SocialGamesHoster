using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Rounds.Endpoints;

public class GetCurrentRound
{
    public record Response(int Id, DateTimeOffset? StartTime, bool IsFinished)
    : IProjectable<Round, Response>
    {
        public int RemainingSeconds { get; init; }
        public bool IsPaused { get; init; }
        public static Expression<Func<Round, Response>> Projection =>
            round => new Response(round.Id, round.StartedTime, round.FinishedTime.HasValue && round.FinishedTime >= DateTimeOffset.UtcNow);
    }

    public static async Task<IResult> HandleAsync(RoundRepository repository, RoundTimer timer, int gameId)
    {
        var roundResult = await repository.GetCurrentRoundFromGame(gameId);
        if (roundResult.IsFailure)
        {
            return roundResult.AsIResult();
        }
        Round round = roundResult.Value;
        TimeSpan timeLeft = timer.RemainingTime;
        bool isFinished = round.FinishedTime.HasValue && round.FinishedTime <= DateTimeOffset.UtcNow;
        Response response = new(round.Id, round.StartedTime, isFinished) { RemainingSeconds = (int)timeLeft.TotalSeconds, IsPaused = timer.State == TimerState.Paused };
        return Results.Ok(response);
    }
}

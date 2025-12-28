using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Rounds.Endpoints;

public class GetCurrentRound
{
    public record Request(int GameId);
    public record Response(int Id, DateTimeOffset? StartTime, bool IsFinished)
    : IProjectable<Round, Response>
    {
        public static Expression<Func<Round, Response>> Projection =>
            round => new Response(round.Id, round.StartedTime, round.FinishedTime.HasValue && round.FinishedTime >= DateTimeOffset.UtcNow);
    }

    public static async Task<IResult> HandleAsync(RoundRepository repository, RoundTimer timer, [AsParameters] Request request)
    {
        var roundResult = await repository.GetCurrentRoundFromGame(request.GameId);
        if (roundResult.IsFailure)
        {
            return roundResult.AsIResult();
        }
        Round round = roundResult.Value;
        TimeSpan timeLeft = timer.RemainingTime;
        bool isFinished = round.FinishedTime.HasValue && round.FinishedTime <= DateTimeOffset.UtcNow;
        Response response = new(round.Id, round.StartedTime, isFinished);
        return Results.Ok(response);
    }
}

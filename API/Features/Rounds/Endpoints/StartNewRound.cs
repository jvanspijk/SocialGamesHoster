using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Rounds.Endpoints;

public static class StartNewRound
{
    public record struct Request(int DurationInSeconds);
    public record Response(int Id, DateTimeOffset? StartTime)
        : IProjectable<Round, Response>
    {
        public int RemainingSeconds { get; init; }
        public static Expression<Func<Round, Response>> Projection =>
            round => new Response(round.Id, round.StartedTime);
    }

    public static async Task<IResult> HandleAsync(RoundRepository repository, RoundTimer timer, int gameId, Request request)
    {
        if (request.DurationInSeconds <= 0)
        {
            return Results.BadRequest("DurationInSeconds must be a positive integer.");
        }
        TimeSpan duration = TimeSpan.FromSeconds(request.DurationInSeconds);
        Round? round = await repository.StartNewRound(gameId);
        if (round == null)
        {
            return Results.NotFound($"Game with id {gameId} not found.");
        }
        timer.Cancel(); // Should we cancel or finish the previous round? Maybe we shouldn't even allow starting a new round if one is active.
        timer.Start(duration, round.Id);

        Response response = new(round.Id, round.StartedTime)
        {
            RemainingSeconds = request.DurationInSeconds
        };

        return Results.Ok(response);
    }
}

using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;

namespace API.Features.Rounds.Endpoints;

public static class StartNewRound
{
    public record struct Request(int DurationInSeconds);

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
        return Results.Created($"/rounds/{round.Id}", round);
    }
}

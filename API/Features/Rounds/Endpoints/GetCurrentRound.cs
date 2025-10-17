using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Features.Rounds.Responses;

namespace API.Features.Rounds.Endpoints;

public class GetCurrentRound
{
    public static async Task<IResult> HandleAsync(RoundRepository repository, RoundTimer timer, int gameId)
    {
        Round? round = await repository.GetCurrentRoundFromGame(gameId);
        if (round == null)
        {
            return Results.NotFound($"Game with id {gameId} does not have an active round or does not exist.");
        }
        TimeSpan timeLeft = timer.RemainingTime;
        if(!round.StartedTime.HasValue)
        {
            return Results.Problem("Round has not started properly.", statusCode: 500);
        }
        RoundResponse response = new(round.Id, round.StartedTime) { RemainingSeconds = (int)timeLeft.TotalSeconds};
        return Results.Ok(response);
    }
}

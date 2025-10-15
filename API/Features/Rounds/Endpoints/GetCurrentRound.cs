using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Rounds.Responses;

namespace API.Features.Rounds.Endpoints;

public class GetCurrentRound
{
    public static async Task<IResult> HandleAsync(RoundRepository repository, int gameId)
    {
        Round? round = await repository.GetCurrentRoundFromGame(gameId);
        if (round == null)
        {
            return Results.NotFound($"Game with id {gameId} does not have an active round or does not exist.");
        }
        RoundResponse response = new(round.Id, round.StartTime, round.EndTime);
        return Results.Ok(response);
    }
}

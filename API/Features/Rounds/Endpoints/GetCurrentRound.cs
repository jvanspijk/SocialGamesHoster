using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Rounds.Responses;

namespace API.Features.Rounds.Endpoints;

public class GetCurrentRound
{
    public static async Task<IResult> HandleAsync(RoundRepository roundService)
    {
        Round? round = await roundService.GetCurrentRound();
        if (round == null)
        {
            return Results.NotFound();
        }
        RoundResponse response = round.ProjectTo<Round, RoundResponse>().First();
        return Results.Ok(response);
    }
}

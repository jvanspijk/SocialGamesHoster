using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Rounds.Responses;

namespace API.Features.Rounds.Endpoints;

public class GetCurrentEndTime
{
    public static async Task<IResult> HandleAsync(RoundRepository roundService)
    {
        Round? round = await roundService.GetCurrentRound();
        if (round == null)
        {
            return Results.NotFound();
        }
        return Results.Ok(new RoundEndTimeResponse() { EndTimeUTC = round.EndTime });
    }
}

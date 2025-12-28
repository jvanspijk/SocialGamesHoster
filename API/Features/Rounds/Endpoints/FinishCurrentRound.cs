using API.Domain;
using API.Features.Rounds.Hubs;
using Microsoft.AspNetCore.SignalR;
using System.Threading.Tasks;

namespace API.Features.Rounds.Endpoints;

public static class FinishCurrentRound
{
    public static async Task<IResult> HandleAsync(IHubContext<RoundsHub, IRoundsHub> hub)
    {
        await RoundsHub.NotifyRoundEnded(hub, 0);
        //TODO: Implement finishing the current round
        return Results.NoContent();
    }
}

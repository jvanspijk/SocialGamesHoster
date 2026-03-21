using API.Domain;
using API.Domain.Models;
using API.Features.GameSessions.Hubs;
using Microsoft.AspNetCore.SignalR;
using System.Threading.Tasks;

namespace API.Features.GameSessions.Endpoints;

public static class FinishCurrentRound
{
    public static async Task<IResult> HandleAsync(IHubContext<GameSessionsHub, IGameSessionsHub> hub)
    {
        await GameSessionsHub.NotifyRoundEnded(hub, 0);
        //TODO: Implement finishing the current round
        return Results.NoContent();
    }
}

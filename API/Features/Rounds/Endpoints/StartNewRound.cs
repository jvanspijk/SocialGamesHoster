using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Rounds.Hubs;
using Microsoft.AspNetCore.SignalR;
using System.Linq.Expressions;

namespace API.Features.Rounds.Endpoints;

public static class StartNewRound
{
    public record Request(int GameId);
    public record Response(int Id)
        : IProjectable<Round, Response>
    {
        public static Expression<Func<Round, Response>> Projection =>
            round => new Response(round.Id);
    }

    public static async Task<IResult> HandleAsync(RoundRepository repository, IHubContext<RoundsHub, IRoundsHub> hub, Request request)
    {        
        Round? round = await repository.StartNewRound(request.GameId);
        if (round == null)
        {
            return Results.NotFound($"Game with id {request.GameId} not found.");
        }

        await RoundsHub.NotifyRoundStarted(hub, round.Id);
        Response response = new(round.Id);
        return Results.Ok(response);
    }
}

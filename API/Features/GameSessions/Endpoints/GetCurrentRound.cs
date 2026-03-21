using API.DataAccess;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public class GetCurrentRound
{
    public record Request(int GameId);
    public record Response(int Id, int RoundNumber, DateTimeOffset? StartTime, Phase? CurrentPhase)
    : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            game => new Response(
                game.Id,
                game.RoundNumber,
                game.RoundStartedAt,
                game.CurrentPhase != null ? new Phase(
                    game.CurrentPhase.Id,
                    game.CurrentPhase.Name,
                    game.CurrentPhase.Description) : null
                );
    }

    public static async Task<IResult> HandleAsync(IRepository<GameSession> repository, [AsParameters] Request request)
    {
        var response = await repository.GetReadOnlyAsync<Response>(r => r.Id == request.GameId);
        if (response is null)
        {
            return Results.NotFound($"Can't find game session with id `{request.GameId}`");
        }

        return Results.Ok(response);
    }
}


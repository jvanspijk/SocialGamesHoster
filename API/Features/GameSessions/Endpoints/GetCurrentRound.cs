using API.DataAccess;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using Microsoft.AspNetCore.Http.HttpResults;
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

    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, [AsParameters] Request request)
    {
        Response? response = await repository.GetReadOnlyAsync<Response>(r => r.Id == request.GameId);
        if (response is null)
        {
            return APIResults.NotFound<GameSession>(request.GameId);
        }

        return APIResults.Ok(response);
    }
}


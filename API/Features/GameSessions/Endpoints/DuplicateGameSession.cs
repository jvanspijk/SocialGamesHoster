using API.DataAccess;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using Microsoft.AspNetCore.Http.HttpResults;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class DuplicateGameSession
{
    public record Request(int GameSessionId);
    public record Response(int GameSessionId, int RulesetId, List<Participant> Participants) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            static gs => new Response(
                gs.Id,
                gs.RulesetId,
                gs.Participants
                    .Select(p => new Participant(
                        p.Id,
                        p.Name,
                        p.Role == null ? null :
                        new RoleInfo(
                            p.Role.Id,
                            p.Role.Name)))
                    .ToList()
            );
    }
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, Request request)
    {
        var gameSession = await repository.GetWithTrackingAsync(request.GameSessionId);
        if (gameSession == null)
        {
            return APIResults.NotFound<GameSession>(request.GameSessionId);
        }

        GameSession duplicatedSession = new()
        {
            RulesetId = gameSession.RulesetId,
            Participants = [.. gameSession.Participants],
            RoundNumber = 0,
            Status = GameStatus.NotStarted
        };

        repository.Add(duplicatedSession);
        await repository.SaveChangesAsync();

        var response = duplicatedSession.ConvertToResponse<GameSession, Response>();
        return APIResults.CreatedAtRoute(response, nameof(GetGameSession), duplicatedSession.Id);
    }
}

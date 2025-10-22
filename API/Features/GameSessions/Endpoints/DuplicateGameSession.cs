using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.GameSessions.Common;
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
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, Request request)
    {
        var gameSession = await repository.GetAsync(request.GameSessionId);
        if (gameSession == null)
        {
            return Results.NotFound($"Game session with id {request.GameSessionId} not found.");
        }

        GameSession duplicatedSession = new()
        {
            RulesetId = gameSession.RulesetId,
            Participants = [.. gameSession.Participants],
            Rounds = [],
            Status = GameStatus.NotStarted
        };

        GameSession createdSession = await repository.CreateAsync(duplicatedSession);

        var response = createdSession.ConvertToResponse<GameSession, Response>();
        return Results.CreatedAtRoute(
            routeName: "GetGameSession",
            routeValues: createdSession.Id,
            value: response);
    }
}

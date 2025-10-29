using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class CreateGameSession
{
    public record struct Request(int RulesetId, List<string> PlayerNames);
    public record Response(int Id, int RulesetId) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(gs.Id, gs.RulesetId);
    }
    public static async Task<IResult> HandleAsync(GameSessionRepository repository, PlayerRepository playerRepository, Request request)
    {
        List<Player> participants = new(request.PlayerNames.Count);
        if (request.PlayerNames.Count > 0)
        {
            participants = await playerRepository.CreateMultipleAsync(
                [.. request.PlayerNames.Select(name => new Player { Name = name })]
            );
        }

        GameSession newSession = new()
        {
            RulesetId = request.RulesetId,
            Participants = participants,
            Rounds = [],
            Status = GameStatus.NotStarted
        };
        Result<GameSession> result = await repository.CreateAsync(newSession);
        if (result.IsFailure)
        {
            return result.AsIResult();
        }
        GameSession session = result.Value;
        Response response = new(session.Id, session.RulesetId);
        return Results.Ok(response);
    }
}

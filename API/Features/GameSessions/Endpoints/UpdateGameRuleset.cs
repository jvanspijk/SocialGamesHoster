using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class UpdateGameRuleset
{
    public readonly record struct Request(int RulesetId);
    public record Response(int Id, int RulesetId) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.RulesetId
            );
    }
    public static async Task<IResult> HandleAsync(
        int gameId,
        Request request,
        IRepository<Ruleset> rulesetRepository,
        IRepository<GameSession> gameSessionRepository)
    {
        var existingGameSession = await gameSessionRepository.GetAsync(gameId);
        if (existingGameSession == null)
        {
            return Results.NotFound();
        }

        if(existingGameSession.Status != GameStatus.NotStarted)
        {
            return Results.BadRequest("Rulesets can only be changed for games that have not been started.");
        }

        if(existingGameSession.RulesetId == request.RulesetId)
        {
            return Results.Ok(existingGameSession.ConvertToResponse<GameSession, Response>());
        }

        var newRuleset = await rulesetRepository.GetAsync(request.RulesetId);
        if (newRuleset == null)
        {
            return Results.NotFound($"Ruleset with id {request.RulesetId} not found.");
        }

        existingGameSession.RulesetId = request.RulesetId;
        await gameSessionRepository.UpdateAsync(existingGameSession);
        var response = existingGameSession.ConvertToResponse<GameSession, Response>();
        return Results.Ok(response);
    }
}

using API.DataAccess;
using API.Domain.Entities;
using API.Domain.Models;
using Microsoft.AspNetCore.Http.HttpResults;
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
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(
        int gameId,
        Request request,
        IRepository<Ruleset> rulesetRepository,
        IRepository<GameSession> gameSessionRepository)
    {
        var existingGameSession = await gameSessionRepository.GetWithTrackingAsync(gameId);
        if (existingGameSession == null)
        {
            return APIResults.NotFound<GameSession>(gameId);
        }

        if(existingGameSession.Status != GameStatus.NotStarted)
        {
            return APIResults.BadRequest("Rulesets can only be changed for games that have not been started.");
        }

        if(existingGameSession.RulesetId == request.RulesetId)
        {
            return APIResults.Ok(existingGameSession.ConvertToResponse<GameSession, Response>());
        }

        var newRuleset = await rulesetRepository.GetWithTrackingAsync(request.RulesetId);
        if (newRuleset == null)
        {
            return APIResults.NotFound($"Ruleset with id {request.RulesetId} not found.");
        }

        existingGameSession.RulesetId = request.RulesetId;
        await gameSessionRepository.SaveChangesAsync();
        var response = existingGameSession.ConvertToResponse<GameSession, Response>();
        return APIResults.Ok(response);
    }
}

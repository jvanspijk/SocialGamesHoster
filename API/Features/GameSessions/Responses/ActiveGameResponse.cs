using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Responses;

public record ActiveGameResponse(int Id, int RulesetId, string Status, int CurrentRoundNumber) : IProjectable<GameSession, ActiveGameResponse>
{
    public static Expression<Func<GameSession, ActiveGameResponse>> Projection => 
        gs => new ActiveGameResponse(
            gs.Id, 
            gs.RulesetId, 
            gs.Status.ToString(),
            gs.Rounds.Count(r => r.StartedTime.HasValue)
        );
}



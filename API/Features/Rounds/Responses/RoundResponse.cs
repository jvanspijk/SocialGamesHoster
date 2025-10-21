using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Rounds.Responses;

public record RoundResponse(int Id, DateTimeOffset? StartTime, bool IsFinished) 
    : IProjectable<Round, RoundResponse>
{
    public int RemainingSeconds { get; init; }
    public static Expression<Func<Round, RoundResponse>> Projection => 
        round => new RoundResponse(round.Id, round.StartedTime, round.FinishedTime.HasValue && round.FinishedTime >= DateTimeOffset.UtcNow);    
}

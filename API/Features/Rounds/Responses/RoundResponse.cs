using API.DataAccess;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Rounds.Responses;

public record RoundResponse(int Id, DateTime StartTime, DateTime EndTime) 
    : IProjectable<Round, RoundResponse>
{
    public static Expression<Func<Round, RoundResponse>> Projection => 
        round => new RoundResponse(round.Id, round.StartTime, round.EndTime);    
}

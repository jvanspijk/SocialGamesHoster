using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints
{
    public static class GetAllGameSessions
    {
        public record Response(int Id, string RulesetName, List<int> ParticipantIds, string Status) : IProjectable<GameSession, Response>
        {
            public static Expression<Func<GameSession, Response>> Projection =>
                gs => new Response(
                    gs.Id,
                    gs.Ruleset!.Name,
                    gs.Participants.Select(p => p.Id).ToList(),
                    gs.Status.ToFriendlyString()
                );
        }

        public static async Task<IResult> HandleAsync(
            GameSessionRepository gameSessionRepository)
        {
            var gameSessions = await gameSessionRepository.GetAllAsync<Response>();
            return Results.Ok(gameSessions);
        }


    }
}

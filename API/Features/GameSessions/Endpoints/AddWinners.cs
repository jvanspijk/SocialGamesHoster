using API.DataAccess;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;
public static class AddWinners
{
    public readonly record struct Request(List<int> PlayerIds);
    public record Response(int Id, List<Participant> Winners) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.Winners
                    .Select(static p => new Participant(
                        p.Id,
                        p.Name,
                        p.Role == null ? null :
                        new RoleInfo(
                            p.Role.Id,
                            p.Role.Name)))
                    .ToList()
            );
    }
    public static async Task<IResult> HandleAsync(
        int gameId,
        Request request,
        IRepository<Player> playerRepository,
        IRepository<GameSession> gameSessionRepository)
    {
        var session = await gameSessionRepository.GetAsync(gameId);
        if (session is null) return TypedResults.NotFound();

        if (session.Status is GameStatus.NotStarted)
        {
            return TypedResults.BadRequest("Winners can only be added to active or finished sessions.");
        }

        var participantIds = session.Participants.Select(p => p.Id).ToHashSet();
        var nonParticipants = request.PlayerIds.Where(id => !participantIds.Contains(id)).ToList();

        if (nonParticipants.Count != 0)
        {
            return TypedResults.BadRequest($"Players {string.Join(", ", nonParticipants)} are not participants.");
        }

        var currentWinnerIds = session.Winners.Select(w => w.Id).ToHashSet();
        var alreadyWinners = request.PlayerIds.Where(id => currentWinnerIds.Contains(id)).ToList();

        if (alreadyWinners.Count != 0)
        {
            return TypedResults.Conflict($"Players {string.Join(", ", alreadyWinners)} are already winners.");
        }

        var playersResult = await playerRepository.GetMultipleAsync(request.PlayerIds);
        if (playersResult.IsFailure)
        {
            return playersResult.AsIResult();
        }

        session.Winners = [.. session.Winners, .. playersResult.Value];

        await gameSessionRepository.UpdateAsync(session);

        var response = await gameSessionRepository.GetAsync<Response>(gameId);
        return TypedResults.Ok(response!);
    }
}

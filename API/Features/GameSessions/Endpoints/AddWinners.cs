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
        var existingGameSession = await gameSessionRepository.GetAsync(gameId);
        if (existingGameSession == null)
        {
            return Results.NotFound();
        }

        if(existingGameSession.Status == GameStatus.NotStarted)
        {
            return Results.BadRequest("Winners can only be added to game sessions that are active or finished.");
        }
        
        List<int> participantIds = [.. existingGameSession.Participants.Select(p => p.Id)];
        var nonParticipants = request.PlayerIds.Except(participantIds).ToList();
        if (nonParticipants.Count > 0)
        {
            return Results.BadRequest($"Players with ids {string.Join(" ,", nonParticipants)} are not participants in this game session.");
        }

        List<int> currentWinnerIds = [.. existingGameSession.Winners.Select(w => w.Id)];
        var alreadyWinners = request.PlayerIds.Intersect(currentWinnerIds).ToList();
        if (alreadyWinners.Count > 0)
        {
            return Results.Conflict($"Players with ids {string.Join(" ,", alreadyWinners)} are already marked as a winner.");
        }

        var players = await playerRepository.GetMultipleAsync(request.PlayerIds);
        if (players.IsFailure)
        {
            return players.AsIResult();
        }

        var newWinners = existingGameSession.Winners.Concat(players.Value).ToList();
        existingGameSession.Winners = newWinners;

        await gameSessionRepository.UpdateAsync(existingGameSession);

        var response = await gameSessionRepository.GetAsync<Response>(gameId);

        return Results.Ok(response);
    }
}

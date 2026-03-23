using API.DataAccess;
using API.Domain.Entities;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using Microsoft.AspNetCore.Http.HttpResults;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;
public static class UpdateGameParticipants
{
    public readonly record struct Request(List<int> ParticipantIds);
    public record Response(int Id, List<Participant> Participants) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(
                gs.Id,
                gs.Participants
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
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(
        int gameId,
        Request request,
        IRepository<Player> playerRepository,
        IRepository<GameSession> gameSessionRepository)
    {
        var existingGameSession = await gameSessionRepository.GetWithTrackingAsync(gameId);
        if (existingGameSession == null)
        {
            return APIResults.NotFound<GameSession>(gameId);
        }  

        if(existingGameSession.Status == GameStatus.Finished)
        {
            return APIResults.BadRequest("Participants cannot be changed when the game is finished.");
        }

        // Participants can only be changed when the game is not running
        // The reason is that changing participants mid-game could lead to inconsistencies, such as removing players
        // For now we only allow changing participants before the game starts
        // This is a stupid rule, this means that if a game is paused, you cannot change participants, which is not ideal, so this needs to be changed.
        if (existingGameSession.Status != GameStatus.NotStarted)
        {
            return APIResults.BadRequest("Participants can only be added to games that have not been started.");
        }

        var newPlayers = await playerRepository.GetListWithTrackingAsync(request.ParticipantIds);
        if(newPlayers.Count != request.ParticipantIds.Count)
        {
            return APIResults.BadRequest("One or more participant IDs are invalid.");
        }

        existingGameSession.Participants = newPlayers;

        await gameSessionRepository.SaveChangesAsync();

        Response response = existingGameSession.ConvertToResponse<GameSession, Response>();
        return APIResults.Ok(response);
    }
}

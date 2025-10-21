using API.DataAccess;
using API.Domain.Models;
using API.Features.GameSessions.Common;
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

        if(existingGameSession.Status == GameStatus.Finished)
        {
            return Results.BadRequest("Participants cannot be changed when the game is finished.");
        }

        // Participants can only be changed when the game is not running
        // The reason is that changing participants mid-game could lead to inconsistencies, such as removing players
        // For now we only allow changing participants before the game starts
        if (existingGameSession.Status != GameStatus.NotStarted)
        {
            return Results.BadRequest("Participants can only be added to games that have not been started.");
        }

        var newPlayers = await playerRepository.GetMultipleAsync(request.ParticipantIds);
        if (newPlayers.IsFailure)
        {
            return newPlayers.AsIResult();
        }
        existingGameSession.Participants = newPlayers.Value;

        await gameSessionRepository.UpdateAsync(existingGameSession);
        return Results.Ok(existingGameSession.ConvertToResponse<GameSession, Response>);
    }
}

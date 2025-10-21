using API.DataAccess;
using API.Domain.Models;
using API.Features.GameSessions.Common;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;
public static class AddWinner
{
    public readonly record struct Request(int PlayerId);
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
        int id,
        Request request,
        IRepository<Player> playerRepository,
        IRepository<GameSession> gameSessionRepository)
    {
        var existingGameSession = await gameSessionRepository.GetAsync(id);
        if (existingGameSession == null)
        {
            return Results.NotFound();
        }

        if(existingGameSession.Status == GameStatus.NotStarted)
        {
            return Results.BadRequest("Winners can only be added to game sessions that are active or finished.");
        }

        if (!existingGameSession.Participants.Any(p => p.Id == request.PlayerId))
        {
            return Results.BadRequest($"Player with id {request.PlayerId} is not a participant in this game session.");
        }

        if (existingGameSession.Winners.Any(w => w.Id == request.PlayerId))
        {
            return Results.BadRequest($"Player with id {request.PlayerId} is already marked as a winner.");
        }

        var player = await playerRepository.GetAsync(request.PlayerId);
        if (player == null)
        {
            return Results.NotFound($"Player with id {request.PlayerId} not found.");
        }        

        existingGameSession.Winners.Add(player);
        await gameSessionRepository.UpdateAsync(existingGameSession);
        var response = existingGameSession.ConvertToResponse<GameSession, Response>();
        return Results.Ok(response);
    }
}

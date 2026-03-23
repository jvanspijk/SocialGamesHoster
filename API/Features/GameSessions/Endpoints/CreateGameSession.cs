using API.DataAccess;
using API.Domain;
using API.Domain.Entities;
using API.Domain.Models;
using Microsoft.AspNetCore.Http.HttpResults;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class CreateGameSession
{
    public record struct Request(int RulesetId, List<string> PlayerNames);
    public record Response(int Id, int RulesetId) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(gs.Id, gs.RulesetId);
    }
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, IRepository<Player> playerRepository, Request request)
    {    
        var participants = playerRepository.AddMultiple([.. request.PlayerNames.Select(name => new Player { Name = name })]);
        GameSession newSession = new()
        {
            RulesetId = request.RulesetId,
            Participants = [.. participants],
            RoundNumber = 0,
            Status = GameStatus.NotStarted
        };

        repository.Add(newSession);
        await repository.SaveChangesAsync();

        Response response = new(newSession.Id, newSession.RulesetId);
        return APIResults.CreatedAtRoute(response, nameof(GetGameSession), newSession.Id);
    }
}

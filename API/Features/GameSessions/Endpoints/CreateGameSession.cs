using API.DataAccess;
using API.Domain;
using API.Domain.Entities;
using API.Domain.Models;
using Microsoft.AspNetCore.Http.HttpResults;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class CreateGameSession
{
    public record struct Request(int RulesetId);
    public record Response(int Id, int RulesetId) : IProjectable<GameSession, Response>
    {
        public static Expression<Func<GameSession, Response>> Projection =>
            gs => new Response(gs.Id, gs.RulesetId);
    }
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(IRepository<GameSession> repository, IRepository<ChatChannel> channelRepository, Request request)
    {          
        GameSession newSession = new()
        {
            RulesetId = request.RulesetId,
            RoundNumber = 0,
            Status = GameStatus.NotStarted
        };

        repository.Add(newSession);
        await repository.SaveChangesAsync();

        channelRepository.Add(new ChatChannel
        {
            Name = "Global",
            GameId = newSession.Id
        });
        await channelRepository.SaveChangesAsync();

        Response response = new(newSession.Id, newSession.RulesetId);
        return APIResults.CreatedAtRoute(response, nameof(GetGameSession), newSession.Id);
    }
}

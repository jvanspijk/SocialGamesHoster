using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
using System.Linq.Expressions;

namespace API.Features.GameSessions.Endpoints;

public static class GetGlobalChat
{
    public record Response(int ChannelId, string ChannelName)
        : IProjectable<ChatChannel, Response>
    {
        public static Expression<Func<ChatChannel, Response>> Projection =>
            channel => new Response(channel.Id, "Global");               
    }

    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(IRepository<ChatChannel> repository, int gameId)
    {
        Response? response = await repository.GetReadOnlyAsync<Response>(c => c.GameId == gameId && c.Name == "Global");
        if(response is null)
        {
            return APIResults.NotFound($"Global chat channel for game with id {gameId} not found.");
        }
        return APIResults.Ok(response);
    }
}

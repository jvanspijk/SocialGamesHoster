using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Chat.Endpoints;

public class CreateChannel
{
    public record Request(string Name, int GameId);
    public record Response(int Id, string Name, int GameId);
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(IRepository<ChatChannel> repository, Request request)
    {
        var channel = new ChatChannel { Name = request.Name, GameId = request.GameId };
        repository.Add(channel);
        await repository.SaveChangesAsync();

        var response = new Response(channel.Id, channel.Name, channel.GameId);
        return APIResults.CreatedAtRoute(response, nameof(GetMessagesFromChannel), channel.Id);
    }
}

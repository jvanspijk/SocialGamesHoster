using API.DataAccess.Repositories;

namespace API.Features.Chat.Endpoints;

public class CreateChannel
{
    public record Request(string Name, int GameId);
    public record Response(int Id, string Name, int GameId);
    public static async Task<IResult> HandleAsync(ChatRepository repository, Request request)
    {
        var channel = await repository.CreateChannelAsync(request.Name, request.GameId);
        var response = new Response(channel.Id, channel.Name, channel.GameId);
        return Results.Ok(response);
    }
}

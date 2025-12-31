using API.DataAccess.Repositories;
using API.Features.Chat.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Chat.Endpoints;

public static class SendMessage
{
    public record Request(int PlayerId, string Message);
    public static async Task<IResult> HandleAsync(ChatRepository chatRepository, IHubContext<ChatHub, IChatHub> hub, int channelId, Request request)
    {
        var result = await chatRepository.CreateAsync(request.PlayerId, channelId, request.Message);
        if(result.IsFailure)
        {
            return result.AsIResult();
        }

        await ChatHub.NotifyMessageSent(hub, channelId, request.PlayerId, result.Value.Id);
        return Results.Created();
    }
}

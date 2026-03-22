using API.DataAccess;
using API.Domain.Entities;
using API.Features.Chat.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Chat.Endpoints;

public static class SendMessage
{
    public record Request(int PlayerId, string Message);
    public static async Task<IResult> HandleAsync(IRepository<ChatMessage> repository, IHubContext<ChatHub, IChatHub> hub, int channelId, Request request)
    {
        bool channelExists = await repository.ExistsAsync(channelId);
        if(!channelExists)
        {
            return Results.NotFound($"Channel with id {channelId} does not exist.");
        }
      
        ChatMessage message = new()
        {
            SenderId = request.PlayerId,
            ChannelId = channelId,
            Content = request.Message,
            IsDeleted = false
        };

        repository.Add(message);
        await repository.SaveChangesAsync();
        await ChatHub.NotifyMessageSent(hub, channelId, request.PlayerId, message.Id);
        return Results.Created();
    }
}

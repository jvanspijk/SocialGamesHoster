using API.DataAccess;
using API.Domain.Entities;
using API.Features.Chat.Hubs;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Chat.Endpoints;

public static class SendMessage
{
    public readonly record struct Request(int PlayerId, string Message);
    public readonly record struct Response(int MessageId, int PlayerId, int ChannelId, string Message);
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(
        IRepository<ChatMessage> messageRepository, IRepository<ChatChannel> channelRepository, 
        IHubContext<ChatHub, IChatHub> hub, int channelId, Request request)
    {
        bool channelExists = await channelRepository.ExistsAsync(channelId);
        if(!channelExists)
        {
            return APIResults.NotFound($"Channel with id {channelId} does not exist.");
        }
      
        ChatMessage message = new()
        {
            SenderId = request.PlayerId,
            ChannelId = channelId,
            Content = request.Message,
            IsDeleted = false
        };

        messageRepository.Add(message);
        await messageRepository.SaveChangesAsync();

        var response = new Response(message.Id, message.SenderId!.Value, message.ChannelId, message.Content);
        await ChatHub.NotifyMessageSent(hub, channelId, request.PlayerId, message.Id);

        return APIResults.CreatedAtRoute(response, nameof(GetMessage), message.Id);
    }
}

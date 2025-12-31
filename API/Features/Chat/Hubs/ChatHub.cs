using Microsoft.AspNetCore.SignalR;

namespace API.Features.Chat.Hubs
{
    public interface IChatHub
    {
        Task ChannelCreated(int channelId, List<int> memberIds, int gameId);
        Task MessageSent(int channelId, int senderId, Guid messageId);
    }
    public class ChatHub : Hub<IChatHub>
    {
        public static async Task NotifyMessageSent(IHubContext<ChatHub, IChatHub> hubContext, int channelId, int senderId, Guid messageId)
        {
            await hubContext.Clients.All.MessageSent(channelId, senderId, messageId);
        }
    }
}

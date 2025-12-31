using Microsoft.AspNetCore.SignalR;

namespace API.Features.Chat.Hubs
{
    public interface IChatHub
    {
        Task OnChannelCreated(int channelId, List<int> memberIds, int gameId);
        Task OnMessageSent(int channelId, int senderId);
    }
    public class ChatHub : Hub<IChatHub>
    {
        public static async Task NotifyMessageSent(IHubContext<ChatHub, IChatHub> hubContext, int channelId, int senderId)
        {
            await hubContext.Clients.All.OnMessageSent(channelId, senderId);
        }
    }
}

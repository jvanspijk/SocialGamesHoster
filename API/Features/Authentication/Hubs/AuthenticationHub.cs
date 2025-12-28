using Microsoft.AspNetCore.SignalR;

namespace API.Features.Authentication.Hubs;

public interface IAuthenticationHub
{
    Task PlayerLoggedIn(int gameId, int playerId);
}

public class AuthenticationHub : Hub<IAuthenticationHub>
{
    public static async Task NotifyPlayerLoggedIn(IHubContext<AuthenticationHub, IAuthenticationHub> hubContext, int gameId, int playerId)
    {
        await hubContext.Clients.All.PlayerLoggedIn(gameId, playerId);
    }
}

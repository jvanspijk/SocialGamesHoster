using Microsoft.AspNetCore.SignalR;

namespace API.Features.Auth.Hubs;

public interface IAuthenticationHub
{
    Task PlayerForceLoggedOut(int gameId, int playerId);
    Task PlayerLoggedOut(int gameId, int playerId);
    Task PlayerLoggedIn(int gameId, int playerId);
}

public class AuthenticationHub : Hub<IAuthenticationHub>
{
    public static async Task NotifyPlayerForceLoggedOut(IHubContext<AuthenticationHub, IAuthenticationHub> hubContext, int gameId, int playerId)
    {
        await hubContext.Clients.All.PlayerForceLoggedOut(gameId, playerId);
    }

    public static async Task NotifyPlayerLoggedOut(IHubContext<AuthenticationHub, IAuthenticationHub> hubContext, int gameId, int playerId)
    {
        await hubContext.Clients.All.PlayerLoggedOut(gameId, playerId);
    }

    public static async Task NotifyPlayerLoggedIn(IHubContext<AuthenticationHub, IAuthenticationHub> hubContext, int gameId, int playerId)
    {
        await hubContext.Clients.All.PlayerLoggedIn(gameId, playerId);
    }
}

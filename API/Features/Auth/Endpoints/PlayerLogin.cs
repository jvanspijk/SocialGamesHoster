using API.DataAccess;
using API.Domain.Entities;
using API.Features.Auth.Hubs;
using Microsoft.AspNetCore.SignalR;
using Microsoft.Extensions.Caching.Memory;

namespace API.Features.Auth.Endpoints;

public static class PlayerLogin
{
    private static string CacheKey(int playerId) => $"{nameof(PlayerLogin)}_{playerId}";
    public readonly record struct Request(int PlayerId, string IPAddress, int GameId);
    public readonly record struct Response(string Token);
    public static async Task<IResult> HandleAsync(AuthService authService, IRepository<Player> playerRepository, IMemoryCache cache, HttpContext httpContext, IHubContext<AuthenticationHub, IAuthenticationHub> hub, Request request)
    {
        Player? player = await cache.GetOrCreateAsync(CacheKey(request.PlayerId), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(10);
            return await playerRepository.GetReadOnlyAsync(r=> r.Id == request.PlayerId);
        }); 
        //Player? player = await playerRepository.GetAsync(request.PlayerId);
        if (player == null)
        {
            return Results.NotFound($"Player with id {request.PlayerId} not found.");
        }
        if(player.GameId == null)
        {
            return Results.BadRequest($"Player with id {request.PlayerId} is not assigned to any game.");
        }
        if (player.GameId != request.GameId)
        {
            return Results.BadRequest($"Player does not belong to game with id {request.GameId}.");
        }
        //IPAddress ipAddress = IPAddress.Parse(httpContext.Request.Headers["X-Forwarded-For"].FirstOrDefault() ?? request.IPAddress);
        string token = authService.GeneratePlayerToken(player.Id, player.Name, player.RoleId);
        await AuthenticationHub.NotifyPlayerLoggedIn(hub, player.GameId.Value, player.Id);
        return Results.Ok(new Response(token));
    }
}

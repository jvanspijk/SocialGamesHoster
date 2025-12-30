using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Auth.Hubs;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Auth.Endpoints;

public static class PlayerLogin
{
    public readonly record struct Request(int PlayerId, string IPAddress, int GameId);
    public readonly record struct Response(string Token);
    public static async Task<IResult> HandleAsync(AuthService authService, PlayerRepository playerRepository, HttpContext httpContext, IHubContext<AuthenticationHub, IAuthenticationHub> hub, Request request)
    {
        Player? player = await playerRepository.GetAsync(request.PlayerId);
        if(player == null)
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

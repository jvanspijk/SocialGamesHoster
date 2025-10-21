using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Net;

namespace API.Features.Authentication.Endpoints;

public static class PlayerLogin
{
    public static Task<IResult> HandleAsync(AuthService authService, PlayerRepository playerRepository, HttpContext httpContext, int gameId, string name)
    {
        // This will be a little harder to implement because of the proxy server
        // To get the IP address of the client, we need to look at the X-Forwarded-For header
        // We need to enable forwarding headers in Program.cs for this to work
        // For that we need to know the proxy server's IP address, which is hardcoded
        // Before this can be implemented, we need to pass the proxy server's IP address to the application using environment variables
        throw new NotImplementedException();
    }
}

using API.Features.Players;
using Microsoft.AspNetCore.SignalR;


namespace API.Hubs;

public class PresenceHub : Hub
{
    private readonly PlayerService _playerService;
    public PresenceHub(PlayerService playerService)
    {
        _playerService = playerService;
    }
    public override async Task OnConnectedAsync()
    {
        //var httpContext = Context.GetHttpContext();

        //string? username = null;
        //if (httpContext != null && httpContext.Request.Query.ContainsKey("username"))
        //{
        //    username = httpContext.Request.Query["username"].ToString();
        //}

        //if (string.IsNullOrEmpty(username))
        //{
        //    Console.WriteLine($"Connection attempted without a username. Disconnecting.");
        //    Context.Abort();
        //    return;
        //}
        //User? user = await _userService.GetUserAsync(username);
        //if (user == null)
        //{
        //    Console.WriteLine($"User {username} does not exist. Disconnecting.");
        //    Context.Abort();
        //    return;
        //}

        //if (_userService.UserIsOnline(username))
        //{
        //    Console.WriteLine($"User '{username}' is already online. Rejecting new connection.");
        //    Context.Abort();
        //    return;
        //}

        //try
        //{
        //    _userService.LogInAsync(username);
        //}
        //catch(Exception ex)
        //{
        //    Console.WriteLine(ex.ToString());
        //    Context.Abort();
        //    return;
        //}

        //Console.WriteLine($"{username} connected.");


        //await Clients.All.SendAsync(
        //    "UserConnectionStatusChanged",
        //    user.Name,
        //    true // isOnline: true
        //);

        await base.OnConnectedAsync();
    }

    public override async Task OnDisconnectedAsync(Exception? exception)
    {
        // TODO:
        //var httpContext = Context.GetHttpContext();
        //string? username = null;
        //if (httpContext != null && httpContext.Request.Query.ContainsKey("username"))
        //{
        //    username = httpContext.Request.Query["username"];
        //}

        //if (!string.IsNullOrEmpty(username))
        //{
        //    var user = _userService.GetUser(username);
        //    if (user != null)
        //    {
        //        _userService.LogOut(username);
        //        Console.WriteLine($"{username} disconnected.");

        //        await Clients.All.SendAsync(
        //            "UserConnectionStatusChanged",
        //            user.Name,
        //            false // isOnline: false
        //        );
        //    }
        //}

        await base.OnDisconnectedAsync(exception);
    }
}

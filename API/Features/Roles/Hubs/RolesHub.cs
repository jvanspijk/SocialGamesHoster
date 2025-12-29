using Microsoft.AspNetCore.SignalR;

namespace API.Features.Roles.Hubs;

public interface IRolesHub
{
    Task RoleUpdated(int roleId);
}

public class RolesHub : Hub<IRolesHub>
{
    public static async Task NotifyRoleUpdated(IHubContext<RolesHub, IRolesHub> hubContext, int roleId)
    {
        await hubContext.Clients.All.RoleUpdated(roleId);
    }
}
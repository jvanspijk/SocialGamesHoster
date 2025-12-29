using Microsoft.AspNetCore.SignalR;

namespace API.Features.Abilities.Hubs;

public interface IAbilitiesHub
{
    Task AbilityUpdated(int abilityId);
}

public class AbilitiesHub : Hub<IAbilitiesHub>
{
    public static async Task NotifyAbilityUpdated(IHubContext<AbilitiesHub, IAbilitiesHub> hubContext, int abilityId)
    {
        await hubContext.Clients.All.AbilityUpdated(abilityId);
    }
}
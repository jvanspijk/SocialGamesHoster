using Microsoft.AspNetCore.SignalR;

namespace API.Features.Rulesets.Hubs;

public interface IRulesetsHub
{
    Task RulesetUpdated(int rulesetId);
}

public class RulesetsHub : Hub<IRulesetsHub>
{
    public static async Task NotifyRulesetUpdated(IHubContext<RulesetsHub, IRulesetsHub> hubContext, int rulesetId)
    {
        await hubContext.Clients.All.RulesetUpdated(rulesetId);
    }
}
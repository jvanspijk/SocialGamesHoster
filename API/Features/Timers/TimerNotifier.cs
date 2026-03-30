using API.Domain;
using API.Features.Timers.Hubs;
using API.Logging;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Timers;

/// <summary>
/// Singleton service that listens for timer finished events and notifies connected clients via the RoundsHub.
/// </summary>
public class TimerNotifier(IGameTimer timer, IHubContext<TimersHub, ITimersHub> hubContext) : IHostedService
{
    private readonly IGameTimer _timer = timer;
    private readonly IHubContext<TimersHub, ITimersHub> _hubContext = hubContext;

    public Task StartAsync(CancellationToken cancellationToken)
    {        
        _timer.OnFinished += HandleTimerFinished;
        return Task.CompletedTask;
    }

    public Task StopAsync(CancellationToken cancellationToken)
    {
        _timer.OnFinished -= HandleTimerFinished;
        return Task.CompletedTask;
    }

    private void HandleTimerFinished()
    {
        Task.Run(async () => await TimersHub.NotifyTimerFinished(_hubContext));
    }
}

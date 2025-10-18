using API.Domain;
using API.Domain.Models;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Rounds.Hubs;

public class TimerEventNotifier : IHostedService
{
    private readonly RoundTimer _timer;
    private readonly IHubContext<TimerHub> _hubContext;

    public TimerEventNotifier(RoundTimer timer, IHubContext<TimerHub> hubContext)
    {
        _timer = timer;
        _hubContext = hubContext;      
    }

    private async void BroadcastTimerStartAsync(TimeSpan remainingTime, int roundId)
    {
        await BroadcastTimeUpdate(remainingTime, isRunning: true, roundId);
    }

    private async void BroadcastTimerPauseAsync(TimeSpan remainingTime, int roundId)
    {
        await BroadcastTimeUpdate(remainingTime, isRunning: false, roundId);
    }

    private async Task BroadcastTimeUpdate(TimeSpan remainingTime, bool isRunning, int roundId)
    {
        await _hubContext.Clients.All.SendAsync("TimerUpdated", new
        {
            RoundId = roundId,
            IsRunning = isRunning,
            RemainingSeconds = remainingTime.TotalSeconds,
        });
    }
    private async void BroadcastTimerCancelledAsync(int roundId)
    {
        await _hubContext.Clients.All.SendAsync("TimerCancelled", roundId);
    }

    private async void BroadcastTimerFinishedAsync(int roundId)
    {
        await _hubContext.Clients.All.SendAsync("TimerFinished", roundId);
    }

    public Task StartAsync(CancellationToken cancellationToken)
    {
        _timer.OnStart += BroadcastTimerStartAsync;
        _timer.OnPause += BroadcastTimerPauseAsync;
        _timer.OnCancel += BroadcastTimerCancelledAsync;
        _timer.OnFinished += BroadcastTimerFinishedAsync;

        return Task.CompletedTask;
    }
    public Task StopAsync(CancellationToken cancellationToken)
    {
        _timer.OnStart -= BroadcastTimerStartAsync;
        _timer.OnPause -= BroadcastTimerPauseAsync;
        _timer.OnCancel -= BroadcastTimerCancelledAsync;
        _timer.OnFinished -= BroadcastTimerFinishedAsync;

        return Task.CompletedTask;
    }
}

using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Rounds.Hubs;

public class TimerEventNotifier(RoundTimer timer, IHubContext<TimerHub> hubContext, IServiceScopeFactory serviceScopeFactory) : IHostedService
{
    private readonly RoundTimer _timer = timer;
    private readonly IHubContext<TimerHub> _hubContext = hubContext;
    private readonly IServiceScopeFactory _serviceScopeFactory = serviceScopeFactory;
    public Task StartAsync(CancellationToken cancellationToken)
    {
        _timer.OnStart += BroadcastTimerStart;

        _timer.OnPause += BroadcastTimerPause;

        _timer.OnCancel += BroadcastTimerCancelled;
        _timer.OnCancel += CancelRoundAsync;

        _timer.OnFinished += BroadcastTimerFinished;
        _timer.OnFinished += FinishRoundAsync;

        return Task.CompletedTask;
    }
    public Task StopAsync(CancellationToken cancellationToken)
    {
        _timer.OnStart -= BroadcastTimerStart;

        _timer.OnPause -= BroadcastTimerPause;

        _timer.OnCancel -= BroadcastTimerCancelled;
        _timer.OnCancel -= CancelRoundAsync;

        _timer.OnFinished -= BroadcastTimerFinished;
        _timer.OnFinished -= FinishRoundAsync;

        return Task.CompletedTask;
    }

    private void BroadcastTimerStart(TimeSpan remainingTime, int roundId)
    {
        BroadcastTimeUpdate(remainingTime, isRunning: true, roundId);
    }

    private void BroadcastTimerPause(TimeSpan remainingTime, int roundId)
    {
       BroadcastTimeUpdate(remainingTime, isRunning: false, roundId);
    }

    private void BroadcastTimeUpdate(TimeSpan remainingTime, bool isRunning, int roundId)
    {
        _hubContext.Clients.All.SendAsync("TimerUpdated", new
        {
            RoundId = roundId,
            IsRunning = isRunning,
            RemainingSeconds = remainingTime.TotalSeconds,
        });
    }
    private void BroadcastTimerCancelled(int roundId)
    {
        _hubContext.Clients.All.SendAsync("TimerCancelled", roundId);
    }

    private void BroadcastTimerFinished(int roundId)
    {
        _hubContext.Clients.All.SendAsync("TimerFinished", roundId);
    }

    private async void FinishRoundAsync(int roundId)
    {
        using var scope = _serviceScopeFactory.CreateScope();
        var roundRepository = scope.ServiceProvider.GetRequiredService<RoundRepository>();
        var result = await roundRepository.FinishRoundAsync(roundId);
        if (result.IsFailure)
        {
            scope.ServiceProvider.GetService<ILogger<TimerEventNotifier>>()?.LogError("Failed to finish round {RoundId}: {Error}", roundId, result.Error.Message);
        }
    }

    private async void CancelRoundAsync(int roundId)
    {
        using var scope = _serviceScopeFactory.CreateScope();
        var roundRepository = scope.ServiceProvider.GetRequiredService<RoundRepository>();
        var result = await roundRepository.CancelRoundAsync(roundId);
        if (result.IsFailure)
        {
            scope.ServiceProvider.GetService<ILogger<TimerEventNotifier>>()?.LogError("Failed to cancel round {RoundId}: {Error}", roundId, result.Error.Message);
        }
    }
}

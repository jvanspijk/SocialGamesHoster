using API.Domain;
using Microsoft.AspNetCore.SignalR;

namespace API.Features.Rounds.Hubs;

public class TimerHub(RoundTimer timer) : Hub
{
    private readonly RoundTimer _timer = timer;

    public Task StartTimer(int seconds, int roundId)
    {
        _timer.Start(TimeSpan.FromSeconds(seconds), roundId);
        return Task.CompletedTask;
    }

    public Task StopTimer()
    {
        _timer.Cancel();
        return Task.CompletedTask;
    }

    public Task PauseTimer()
    {
        _timer.Pause();
        return Task.CompletedTask;
    }

    public Task ResumeTimer()
    {
        _timer.Resume();
        return Task.CompletedTask;
    }

    public Task FinishTimer()
    {
        _timer.Finish();
        return Task.CompletedTask;
    }

    public Task AdjustTimer(int seconds)
    {
        _timer.AdjustTime(TimeSpan.FromSeconds(seconds));
        return Task.CompletedTask;
    }
}

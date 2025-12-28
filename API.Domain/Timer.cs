using API.Domain.Models;
using System.Diagnostics;

namespace API.Domain;

public enum TimerState
{
    Stopped,
    Running,
    Paused
}
public class RoundTimer
{
    private readonly Stopwatch _stopwatch = new();
    private CancellationTokenSource? _timerCts;
    private readonly object _lock = new();
    private TimeSpan _totalTime;

    public TimerState State { get; private set; } = TimerState.Stopped;
    public event Action? OnFinished;

    public TimeSpan TotalTime
    {
        get { lock (_lock) return _totalTime; }
    }

    public TimeSpan ElapsedTime
    {
        get { lock (_lock) return _stopwatch.Elapsed; }
    }

    public TimeSpan RemainingTime
    {
        get
        {
            lock (_lock)
            {
                var remaining = _totalTime - _stopwatch.Elapsed;
                return remaining > TimeSpan.Zero ? remaining : TimeSpan.Zero;
            }
        }
    }

    public void Start(TimeSpan duration)
    {
        lock (_lock)
        {
            _totalTime = duration;
            _stopwatch.Restart();
            State = TimerState.Running;
            UpdateTimerCallback(RemainingTime);
        }
    }

    public void Pause()
    {
        lock (_lock)
        {
            if (State != TimerState.Running) return;
            _stopwatch.Stop();
            _timerCts?.Cancel();
            State = TimerState.Paused;
        }
    }

    public void Resume()
    {
        lock (_lock)
        {
            if (State != TimerState.Paused) return;
            _stopwatch.Start();
            State = TimerState.Running;
            UpdateTimerCallback(RemainingTime);
        }
    }

    public void Stop()
    {
        lock (_lock)
        {
            _timerCts?.Cancel();
            _timerCts?.Dispose();
            _timerCts = null;

            _stopwatch.Reset();
            _totalTime = TimeSpan.Zero;
            State = TimerState.Stopped;
        }
    }

    public void AdjustTime(TimeSpan delta)
    {
        lock (_lock)
        {
            if (State == TimerState.Stopped) return;

            _totalTime += delta;

            if (State == TimerState.Running)
            {
                UpdateTimerCallback(RemainingTime);
            }
        }
    }

    private void UpdateTimerCallback(TimeSpan delay)
    {
        _timerCts?.Cancel();
        _timerCts?.Dispose();

        if (delay <= TimeSpan.Zero)
        {
            HandleTimerFinished();
            return;
        }

        _timerCts = new CancellationTokenSource();
        var token = _timerCts.Token;

        _timerCts.CancelAfter(delay);
        token.Register(() =>
        {
            if (!token.IsCancellationRequested)
            {
                HandleTimerFinished();
            }
        });
    }

    private void HandleTimerFinished()
    {
        lock (_lock)
        {
            State = TimerState.Stopped;
            _stopwatch.Stop();
        }
        OnFinished?.Invoke();
    }
}
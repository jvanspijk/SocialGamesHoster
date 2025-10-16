using API.Domain.Models;

namespace API.Domain;

public enum TimerState
{
    Stopped,
    Running,
    Paused
}

public sealed class RoundTimer
{
    public event Action<TimeSpan, int>? OnStart;
    public event Action<TimeSpan, int>? OnPause;
    public event Action<TimeSpan, int>? OnResume;
    public event Action<TimeSpan, int>? OnAdjust;
    public event Action<int>? OnCancel;
    public event Action<int>? OnFinished;
    public TimeSpan RemainingTime
    {
        get
        {
            lock (_lock)
            {
                if (_state == TimerState.Stopped) return TimeSpan.Zero;
                if (_state == TimerState.Paused) return _pausedRemaining;

                var remaining = _finishTime - DateTimeOffset.UtcNow;
                return remaining > TimeSpan.Zero ? remaining : TimeSpan.Zero;
            }
        }
    }
    public int RoundId
    {
        get
        {
            lock (_lock)
            {
                return _roundId;
            }
        }
        set
        {
            lock (_lock)
            {
                _roundId = value;
            }
        }
    }
    public TimerState State
    {
        get
        {
            lock (_lock)
            {
                return _state;
            }
        }
    }

    private TimerState _state = TimerState.Stopped;
    private DateTimeOffset _finishTime = DateTimeOffset.MinValue;
    private TimeSpan _pausedRemaining = TimeSpan.Zero;

    private CancellationTokenSource? _cts;
    private int _roundId;
    private readonly Lock _lock = new();

    public void Start(TimeSpan duration, int roundId)
    {
        TimeSpan currentRemaining;
        lock (_lock)
        {
            if (duration <= TimeSpan.Zero) 
            { 
                throw new InvalidOperationException("Duration must be positive."); 
            }
            if (_state != TimerState.Stopped) { return; }      

            _pausedRemaining = TimeSpan.Zero;
            _finishTime = DateTimeOffset.UtcNow + duration;
            _state = TimerState.Running;
            _roundId = roundId;

            currentRemaining = RemainingTime;
            CreateCountDownTask_RequiresLock(currentRemaining);
        }
        OnStart?.Invoke(currentRemaining, roundId);
    }

    public void Resume()
    {
        TimeSpan currentRemaining;
        int roundId;

        lock (_lock)
        {
            if (_state != TimerState.Paused) { return; }

            _finishTime = DateTimeOffset.UtcNow + _pausedRemaining;
            _pausedRemaining = TimeSpan.Zero;
            _state = TimerState.Running;

            currentRemaining = RemainingTime;
            CreateCountDownTask_RequiresLock(currentRemaining);
            roundId = _roundId;
        }
        OnResume?.Invoke(currentRemaining, roundId);
    }

    public void Pause()
    {
        TimeSpan currentRemaining;
        CancellationTokenSource? localCts = null;
        int roundId;

        lock (_lock)
        {
            if (_state != TimerState.Running) { return; }

            _pausedRemaining = RemainingTime;
            _finishTime = DateTimeOffset.MinValue;
            _state = TimerState.Paused;

            currentRemaining = _pausedRemaining;
            localCts = _cts;
            roundId = _roundId;
        }

        try 
        { 
            localCts?.Cancel(); 
            localCts?.Dispose(); 
        }
        catch (ObjectDisposedException) { }

        OnPause?.Invoke(currentRemaining, roundId);
    }

    public void Cancel()
    {
        int roundId;
        lock (_lock)
        {
            if (_state == TimerState.Stopped) { return; }
            Reset_RequiresLock();
            roundId = _roundId;
        }
        OnCancel?.Invoke(roundId);
    }

    public void Finish()
    {
        int roundId;
        lock (_lock)
        {
            if (_state == TimerState.Stopped) { return; }
            Reset_RequiresLock();
            roundId = _roundId;
        }
        OnFinished?.Invoke(roundId);
    }

    public void AdjustTime(TimeSpan delta)
    {
        TimeSpan newRemaining;
        int roundId;
        lock (_lock)
        {
            if (_state == TimerState.Stopped) { return; }

            newRemaining = RemainingTime + delta;
            newRemaining = newRemaining > TimeSpan.Zero ? newRemaining : TimeSpan.Zero;

            if (_state == TimerState.Paused)
            {
                _pausedRemaining = newRemaining;
            }
            else if (_state == TimerState.Running)
            {
                _finishTime = DateTimeOffset.UtcNow + newRemaining;
                CreateCountDownTask_RequiresLock(newRemaining);
            }
            roundId = _roundId;
        }

        if (newRemaining == TimeSpan.Zero)
        {
            Finish();
        }
        OnAdjust?.Invoke(newRemaining, roundId);
    }

    private void Reset_RequiresLock()
    {
        _finishTime = DateTimeOffset.MinValue;
        _pausedRemaining = TimeSpan.Zero;
        _state = TimerState.Stopped;

        _cts?.Cancel();
        _cts?.Dispose();
        _cts = null;
    }

    private void CreateCountDownTask_RequiresLock(TimeSpan duration)
    {
        _cts?.Dispose();
        _cts = new CancellationTokenSource();
        var token = _cts.Token;

        if (duration > TimeSpan.Zero)
        {
            _ = CountdownTask(duration, token);
        }
        else
        {
            Reset_RequiresLock();
        }
    }

    private async Task CountdownTask(TimeSpan duration, CancellationToken token)
    {
        try
        {
            await Task.Delay(duration, token);
            Finish();
        }
        catch (TaskCanceledException)
        {
            // Expected on Pause, Stop, or AdjustTime.
        }
    }
}
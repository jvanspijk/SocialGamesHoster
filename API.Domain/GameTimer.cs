namespace API.Domain;

public class GameTimer : IGameTimer
{
    private readonly Lock _lock = new();
    private CancellationTokenSource? _cts;

    private DateTime? _endAtUtc;
    private TimeSpan? _pausedTimeLeft;
    private TimeSpan _totalDuration;

    public event Action? OnFinished;

    public TimerState CurrentState
    {
        get
        {
            lock (_lock)
            {
                if (_endAtUtc.HasValue) return TimerState.Running;
                if (_pausedTimeLeft.HasValue) return TimerState.Paused;
                return _totalDuration > TimeSpan.Zero ? TimerState.Completed : TimerState.Completed;
            }
        }
    }

    public TimeSpan RemainingTime
    {
        get
        {
            lock (_lock)
            {
                if (_pausedTimeLeft.HasValue) return _pausedTimeLeft.Value;
                if (!_endAtUtc.HasValue) return TimeSpan.Zero;

                var remaining = _endAtUtc.Value - DateTime.UtcNow;
                return remaining > TimeSpan.Zero ? remaining : TimeSpan.Zero;
            }
        }
    }

    public TimeSpan TotalTime
    {
        get
        {
            lock (_lock) return _totalDuration;
        }
    }

    public void Start(TimeSpan duration)
    {
        lock (_lock)
        {
            _totalDuration = duration;
            _pausedTimeLeft = null;
            _endAtUtc = DateTime.UtcNow.Add(duration);

            Reschedule(duration);
        }
    }

    public void Pause()
    {
        lock (_lock)
        {
            if (CurrentState != TimerState.Running || !_endAtUtc.HasValue) return;

            _pausedTimeLeft = RemainingTime;
            CancelInternal();
            _endAtUtc = null;
        }
    }

    public void Resume()
    {
        lock (_lock)
        {
            if (CurrentState != TimerState.Paused || !_pausedTimeLeft.HasValue) return;

            var remaining = _pausedTimeLeft.Value;
            _endAtUtc = DateTime.UtcNow.Add(remaining);
            _pausedTimeLeft = null;

            Reschedule(remaining);
        }
    }

    public void Adjust(TimeSpan delta)
    {
        lock (_lock)
        {
            var state = CurrentState;
            if (state == TimerState.Inactive) return;

            if (state == TimerState.Paused || state == TimerState.Completed)
            {
                var newRemaining = RemainingTime + delta;
                _pausedTimeLeft = newRemaining > TimeSpan.Zero ? newRemaining : TimeSpan.Zero;
            }
            else
            {
                _endAtUtc = _endAtUtc?.Add(delta);
                Reschedule(RemainingTime);
            }

            _totalDuration += delta;

            if (_totalDuration < RemainingTime) _totalDuration = RemainingTime;
            if (_totalDuration < TimeSpan.Zero) _totalDuration = TimeSpan.Zero;
        }
    }

    public void Stop()
    {
        lock (_lock)
        {
            CancelInternal();
            _endAtUtc = null;
            _pausedTimeLeft = null;
            _totalDuration = TimeSpan.Zero;
        }
    }

    public void Reset() => Stop();

    private void Reschedule(TimeSpan delay)
    {
        CancelInternal();

        if (delay <= TimeSpan.Zero)
        {
            _ = Task.Run(HandleTimerFinished);
            return;
        }

        _cts = new CancellationTokenSource();
        var token = _cts.Token;

        Task.Delay(delay, token).ContinueWith(t =>
        {
            if (t.IsCompletedSuccessfully) HandleTimerFinished();
        }, TaskContinuationOptions.ExecuteSynchronously);
    }

    private void CancelInternal()
    {
        _cts?.Cancel();
        _cts?.Dispose();
        _cts = null;
    }

    private void HandleTimerFinished()
    {
        lock (_lock)
        {
            if (CurrentState != TimerState.Running) return;

            _endAtUtc = null;
            CancelInternal();
        }
        OnFinished?.Invoke();
    }
}
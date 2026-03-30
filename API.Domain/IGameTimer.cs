namespace API.Domain;

public interface IGameTimer
{
    TimeSpan RemainingTime { get; }
    TimeSpan TotalTime { get; }
    TimerState CurrentState { get; }

    event Action? OnFinished;

    void Adjust(TimeSpan delta);
    void Pause();
    void Resume();
    void Start(TimeSpan duration);
    void Stop();
    void Reset();
}
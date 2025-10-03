namespace API.Domain.Models;

public class Round
{
    public Round(DateTime startTimeUtc, TimeSpan duration)
    {
        StartTime = startTimeUtc;
        EndTime = startTimeUtc.Add(duration);
    }

    public int Id { get; set; }

    /// <summary>
    /// The order of the round in the game session. E.g. 1 for the first round, 2 for the second, etc.
    /// </summary>
    public int Order { get; set; }

    /// <summary>
    /// The time the round started in UTC.
    /// </summary>
    public DateTime StartTime {  get; set; }
    /// <summary>
    /// The time the round is scheduled to end in UTC.
    /// </summary>
    public DateTime EndTime { get; private set; }
    public bool RoundOver {  get => EndTime >= DateTime.UtcNow; }

    public TimeSpan RemainingTime
    {
        get
        {
            var timeLeft = EndTime - DateTime.UtcNow;
            return timeLeft > TimeSpan.Zero ? timeLeft : TimeSpan.Zero;
        }
    }

    public int RemainingSeconds
    {
        get => (int)Math.Ceiling(RemainingTime.TotalSeconds);          
    }
}

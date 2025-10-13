using System.Text.Json.Serialization;

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
    /// The time the round started in UTC.
    /// </summary>
    public DateTime StartTime {  get; set; }
    /// <summary>
    /// The time the round is scheduled to end in UTC.
    /// </summary>
    public DateTime EndTime { get; set; }
    public bool RoundOver {  get => EndTime >= DateTime.UtcNow; }

    public int RoundNumber { get; set; }
    [JsonIgnore]
    public GameSession GameSession { get; set; }
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

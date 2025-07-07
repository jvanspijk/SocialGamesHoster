namespace SocialGamesHoster.API.Models;

public class Round
{
    public Round(DateTime startTimeUtc, TimeSpan duration)
    {
        StartTime = startTimeUtc;
        EndTime = startTimeUtc.Add(duration);
    }
    public DateTime StartTime {  get; set; }
    public DateTime EndTime { get; set; }
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

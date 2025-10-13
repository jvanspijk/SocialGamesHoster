using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace API.Domain.Models;

public enum GameStatus
{
    Unknown = 0,
    Running = 100,
    Paused = 200,
    Finished = 300,
}
public class GameSession
{
    public int Id { get; set; }
    public Ruleset? Ruleset { get; set; }
    [JsonIgnore]
    [ForeignKey(nameof(Ruleset))]
    public required int RulesetId { get; set; }
    public ICollection<Player> Participants { get; set; } = [];
    public ICollection<Round> Rounds { get; set; } = [];
    public Round? CurrentRound { get; set; }
    public ICollection<Player> Winners { get; set; } = [];
    public GameStatus Status { get; set; } = GameStatus.Unknown;
    public void StartNewRound(DateTime startTimeUtc, TimeSpan duration)
    {
        var newRound = new Round(startTimeUtc, duration)
        {
            GameId = Id,
            RoundNumber = (Rounds.Count > 0) ? Rounds.Max(r => r.RoundNumber) + 1 : 1
        };
        Rounds.Add(newRound);
        CurrentRound = newRound;
    }

    public void EndCurrentRound()
    {
        if (CurrentRound != null)
        {
            CurrentRound.EndTime = DateTime.UtcNow;
            CurrentRound = null;
        }
    }
}

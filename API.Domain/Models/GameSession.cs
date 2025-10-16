using System.ComponentModel.DataAnnotations;
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

    public int CurrentRoundId { get; set; }
    public Round? CurrentRound { get; set; }
    public int CurrentRoundNumber => Rounds.Where(r => r.StartedTime != default).Count();
    public ICollection<Player> Winners { get; set; } = [];
    public GameStatus Status { get; set; } = GameStatus.Unknown;       
}

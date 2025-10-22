using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace API.Domain.Models;

public enum GameStatus
{
    NotStarted = 0,
    Running = 100,
    Paused = 200,
    Cancelled = 300,
    Finished = 400,
}
public class GameSession
{
    public int Id { get; set; }
    public Ruleset? Ruleset { get; set; }
    public required int RulesetId { get; set; }
    public ICollection<Player> Participants { get; set; } = [];
    public ICollection<Round> Rounds { get; set; } = [];
    public int CurrentRoundId { get; set; }
    public Round? CurrentRound { get; set; }
    public ICollection<Player> Winners { get; set; } = [];
    public GameStatus Status { get; set; } = GameStatus.NotStarted;

    [NotMapped]
    public bool IsDone => Status == GameStatus.Finished || Status == GameStatus.Cancelled;
    [NotMapped]
    public bool IsActive => Status == GameStatus.Running || Status == GameStatus.Paused;
}

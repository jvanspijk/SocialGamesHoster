using API.Domain.Entities;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.RegularExpressions;

namespace API.Domain.Models;

public enum GameStatus
{
    NotStarted = 0,
    Running = 100,
    Paused = 200,
    Finished = 300,
}
public static partial class GameStatusExtensions
{
    [GeneratedRegex("([a-z])([A-Z])")]
    private static partial Regex SplitCamelCase();

    public static string ToFriendlyString(this GameStatus status) =>
        SplitCamelCase().Replace(status.ToString(), "$1 $2");
}

public class GameSession : IEntity
{
    public int Id { get; set; }
    public Ruleset? Ruleset { get; set; }
    public required int RulesetId { get; set; }
    public ICollection<Player> Participants { get; set; } = [];

    // Round information
    public int RoundNumber { get; set; } = 0;
    public DateTimeOffset RoundStartedAt { get; set; } = DateTimeOffset.UtcNow;

    // Phase information
    public int? CurrentPhaseId { get; set; }
    public RoundPhase? CurrentPhase { get; set; }
    public DateTimeOffset PhaseStartedAt { get; set; } = DateTimeOffset.UtcNow;

    public ICollection<Player> Winners { get; set; } = [];
    public GameStatus Status { get; set; } = GameStatus.NotStarted;
    public ICollection<ChatChannel> ChatChannels { get; set; } = [];

    [NotMapped]
    public bool IsDone => Status == GameStatus.Finished;
    [NotMapped]
    public bool IsActive => Status == GameStatus.Running || Status == GameStatus.Paused;
}

namespace API.Domain.Models;

public class Game
{
    public int Id { get; set; }
    public required Ruleset Ruleset { get; set; }
    public ICollection<Player> Players { get; set; } = [];
    public ICollection<Round> Rounds { get; set; } = [];
    public Round? CurrentRound { get; set; }
}

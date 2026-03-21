namespace API.Domain.Entities;

public class RoundPhase : IEntity
{
    public int Id { get; set; }
    public int RulesetId { get; set; }
    public required string Name { get; set; }
    public string Description { get; set; } = string.Empty;
    public int Order { get; set; }
}
namespace API.Domain.Models;

public enum KnowledgeType
{
    Role = 100,
    Alignment = 200,
}

/// <summary>
/// Many-to-many join table used by Entity Framework.
/// Not intended to be used directly.
/// Source knows KnowledgeType about Target.
/// </summary>
public class RoleKnowledge
{
    public KnowledgeType KnowledgeType { get; set; }
    public int SourceId { get; set; }
    /// <summary>
    /// The knower role
    /// </summary>
    public Role Source { get; set; } = null!;
    public int TargetId { get; set; }
    /// <summary>
    /// The role that has their information exposed
    /// </summary>
    public Role Target { get; set; } = null!;

    public override string ToString()
    {
        return $"{Source.Name} knows {KnowledgeType} from {Target.Name}";
    }
}

namespace API.Models;

[Serializable]
public class Role
{
    public required int Id { get; init; }
    public required string Name { get; set; }
    public string? Description { get; set; }
    public List<Ability> Abilities { get; set; } = [];
}

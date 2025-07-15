using System.Text.Json.Serialization;

namespace API.Models;

[Serializable]
public class Role
{
    public Role()
    {
        Abilities = new List<Ability>();
    }
    public required int Id { get; set; }
    public required string Name { get; set; }
    public string? Description { get; set; }
    public List<Ability> Abilities { get; set; }
}

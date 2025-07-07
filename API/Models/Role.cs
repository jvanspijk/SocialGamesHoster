using System.Text.Json.Serialization;

namespace API.Models;

[Serializable]
public class Role
{
    public required string Name { get; set; } // We assume names are unique
    public string? Description { get; set; }
}

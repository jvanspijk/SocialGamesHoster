using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace API.Models;

public class Ability
{
    [Key]
    public int Id { get; init; }
    public required string Name { get; set; }
    public required string Description { get; set; }
    [JsonIgnore]
    public ICollection<RoleAbility> RoleAssociations { get; set; } = [];
}

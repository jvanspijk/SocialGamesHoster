using System.ComponentModel.DataAnnotations;

namespace API.Models;

public class Ability
{
    [Key]
    public int Id { get; init; }
    public required string Name { get; set; }
    public required string Description { get; set; }
    public ICollection<RoleAbility> RoleAssociations { get; set; } = [];
}

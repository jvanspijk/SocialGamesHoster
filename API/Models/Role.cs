using System.ComponentModel.DataAnnotations;

namespace API.Models;

public class Role
{
    [Key]
    public required int Id { get; set; }
    public required string Name { get; set; }
    public string? Description { get; set; }
    public ICollection<RoleAbility> AbilityAssociations { get; set; } = [];
    // public ICollection<Ability> Abilities => AbilityAssociations.Select(ra => ra.Ability).ToList();
}


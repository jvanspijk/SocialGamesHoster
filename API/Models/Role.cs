using System.ComponentModel.DataAnnotations;

namespace API.Models;

public class Role
{
    [Key]
    public required int Id { get; set; }
    [Required]
    [StringLength(32, MinimumLength = 3, ErrorMessage = "Role name must be between 3 and 32 characters long.")]
    public required string Name { get; set; }
    [StringLength(512)]
    public string? Description { get; set; }
    public ICollection<RoleAbility> AbilityAssociations { get; set; } = [];
    // public ICollection<Ability> Abilities => AbilityAssociations.Select(ra => ra.Ability).ToList();
    public List<int> RolesVisibleToRole { get; set; } = [];
}


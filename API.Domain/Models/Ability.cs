using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;

namespace API.Domain.Models;

public class Ability
{
    [Key]
    public int Id { get; init; }
    [Required]
    [StringLength(32, MinimumLength = 1, ErrorMessage = "Ability name must be between 1 and 32 characters long.")]
    public required string Name { get; set; }
    public required string Description { get; set; }
    public ICollection<Role> AssociatedRoles { get; set; } = [];
    [ForeignKey(nameof(Ruleset))]
    public int RulesetId { get; set; }
}

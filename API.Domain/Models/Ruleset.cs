using System.ComponentModel;
using System.ComponentModel.DataAnnotations;

namespace API.Domain.Models;

public class Ruleset
{
    public int Id { get; init; }
    [Required]
    [StringLength(32, MinimumLength = 1, ErrorMessage = "Game name must be between 1 and 64 characters long.")]
    public required string Name { get; set; }
    [StringLength(512)]
    [DefaultValue("")]
    public string Description { get; set; } = "";
    public ICollection<Role> Roles { get; set; } = [];
    public ICollection<Ability> Abilities { get; set; } = [];
}

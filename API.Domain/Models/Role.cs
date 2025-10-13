using System.ComponentModel;
using System.ComponentModel.DataAnnotations;
using System.ComponentModel.DataAnnotations.Schema;
using System.Text.Json.Serialization;

namespace API.Domain.Models;

public class Role
{
    [Key]
    public int Id { get; set; }
    [Required]
    [StringLength(32, MinimumLength = 2, ErrorMessage = "Role name must be between 2 and 32 characters long.")]
    public required string Name { get; set; }
    [StringLength(512)]
    [DefaultValue("")]
    public string Description { get; set; } = "";
    public ICollection<Ability> Abilities { get; set; } = [];
    public ICollection<Role> CanSee { get; set; } = [];
    public ICollection<Role> CanBeSeenBy { get; set; } = [];
    public ICollection<Player> PlayersWithRole { get; set; } = [];
    [JsonIgnore]
    [ForeignKey(nameof(Ruleset))]
    public int RulesetId { get; set; }
}


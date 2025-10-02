using System.ComponentModel;
using System.ComponentModel.DataAnnotations;

namespace API.Domain.Models;

public class Role
{
    [Key]
    public int Id { get; set; }
    [Required]
    [StringLength(32, MinimumLength = 3, ErrorMessage = "Role name must be between 3 and 32 characters long.")]
    public required string Name { get; set; }
    [StringLength(512)]
    [DefaultValue("")]
    public string Description { get; set; } = "";
    public ICollection<RoleAbility> AbilityAssociations { get; set; } = [];
    public ICollection<RoleVisibility> CanSee { get; set; } = new List<RoleVisibility>();
    public ICollection<RoleVisibility> CanBeSeenBy { get; set; } = new List<RoleVisibility>();
    public ICollection<Player> PlayersWithRole { get; set; } = new List<Player>();
}


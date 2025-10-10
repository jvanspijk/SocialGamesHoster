using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace API.Domain.Models;
//TODO we should use DTO's so that we can hide some fields for players but not for admins
public class Player
{
    [Key]
    [Required]
    public int Id { get; set; }
    [Required]
    [StringLength(64, MinimumLength = 1, ErrorMessage = "Player name must be between 1 and 64 characters long.")]
    [RegularExpression(@"^[a-zA-Z0-9 ]+$", ErrorMessage = "Player name can only contain alphanumeric characters and spaces.")]
    public required string Name { get; set; }

    // instead of making the rule nullable, there should be a "no role" role
    // this way we can avoid null checks everywhere
    public Role? Role { get; set; }

    /// <summary>
    /// A role that the player is disguised as, but don't actually have.
    /// </summary>
    //[JsonIgnore]
    //public Role? DisguisedRole { get; set; }

    [JsonIgnore]
    public int? RoleId { get; set; }
    [JsonIgnore]

    public ICollection<Player> CanSee { get; set; } = [];
    public ICollection<Player> CanBeSeenBy { get; set; } = [];
    public bool IsEliminated { get; set; } = false;
}



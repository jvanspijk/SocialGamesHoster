
using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace API.Models;
//TODO we should use DTO's so that we can hide some fields for players but not for admins
public class Player
{
    [Key]
    [Required]
    public int Id { get; set; }
    [Required]
    [StringLength(32, MinimumLength = 1, ErrorMessage = "Player name must be between 1 and 32 characters long.")]
    [RegularExpression(@"^[a-zA-Z0-9 ]+$", ErrorMessage = "Player name can only contain alphanumeric characters and spaces.")]
    public required string Name { get; set; }

    public Role? Role { get; set; }

    /// <summary>
    /// A role that the player is disguised as, but don't actually have.
    /// </summary>
    //[JsonIgnore]
    //public Role? DisguisedRole { get; set; }

    [JsonIgnore]
    public int? RoleId { get; set; }
    [JsonIgnore]

    public ICollection<PlayerVisibility> CanSee { get; set; } = [];
    public ICollection<PlayerVisibility> CanBeSeenBy { get; set; } = [];
}




using API.Validation;
using System.ComponentModel.DataAnnotations;
using System.Text.Json.Serialization;

namespace API.Models;

public class Player
{
    [Key]
    [Required]
    public int Id { get; set; }
    [Required]
    public required string Name { get; set; }

    [JsonIgnore]
    public Role? Role { get; set; }

    [JsonIgnore]
    public int? RoleId { get; set; } // This is used by EF to set the foreign key relationship

    public static string NormalizeName(string userName)
    {
        return userName.ToLowerInvariant().Normalize();
    }
}



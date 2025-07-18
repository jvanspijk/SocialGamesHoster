using System.ComponentModel.DataAnnotations;

namespace API.DTO;

public class AdminLoginRequestDTO
{
    [Required]
    public required string Username { get; set; }

    [Required]
    public required string Password { get; set; }
}

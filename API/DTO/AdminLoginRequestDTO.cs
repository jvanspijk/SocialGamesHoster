using System.ComponentModel.DataAnnotations;

namespace API.DTO;

public readonly record struct AdminLoginRequestDTO(string Username, string Password);


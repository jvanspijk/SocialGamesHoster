namespace API.Models;

public class User
{
    public required string Name { get; set; } // We assume names are unique
    public Role? Role { get; set; }
}

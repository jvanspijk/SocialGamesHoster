namespace API.Models;

public class RoleVisibility
{
    public int RoleId { get; set; }
    public Role Role { get; set; } = null!;

    public int VisibleRoleId { get; set; }
    public Role VisibleRole { get; set; } = null!;
}
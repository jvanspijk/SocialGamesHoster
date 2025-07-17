namespace API.Models;
// Collection of many-to-many classes needed for EF
public class RoleAbility
{
    public int RoleId { get; init; }
    public Role? Role { get; set; }

    public int AbilityId { get; init; }
    public Ability? Ability { get; set; }
}

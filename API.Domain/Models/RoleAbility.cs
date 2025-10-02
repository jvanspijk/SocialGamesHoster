namespace API.Domain.Models;
public class RoleAbility
{
    public int RoleId { get; init; }
    public Role? Role { get; set; }

    public int AbilityId { get; init; }
    public Ability? Ability { get; set; }
}

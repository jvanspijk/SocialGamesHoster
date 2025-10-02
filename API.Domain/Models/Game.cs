namespace API.Domain.Models;

public class Game
{
    public int Id { get; set; }
    public string Name { get; set; }
    public string Description { get; set; }
    public ICollection<Role> Roles { get; set; }
    public ICollection<Ability> Abilities { get; set; }
}

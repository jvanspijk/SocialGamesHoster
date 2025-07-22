namespace API.Models;

public class PlayerVisibility
{
    public int PlayerId { get; set; }
    public Player Player { get; set; } = null!;

    public int VisiblePlayerId { get; set; }
    public Player VisiblePlayer { get; set; } = null!;
}
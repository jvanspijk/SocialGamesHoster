namespace API.Domain.Entities;
public class ChatChannel : IEntity
{
    public int Id { get; set; }
    public string Name { get; set; } = string.Empty;
    public int GameId { get; set; }
    public ICollection<ChatMessage> Messages { get; set; } = [];
    public ICollection<ChatChannelMembership> Members { get; set; } = [];
}

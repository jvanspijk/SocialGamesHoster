namespace API.Domain.Entities;

public class ChatMessage : IEntity
{
    public int Id { get; init; }
    public required string Content { get; set; }
    public DateTime SentAt { get; init; } = DateTime.UtcNow;
    public Player? Sender { get; set; }
    public int? SenderId { get; set; }
    public bool IsAdmin { get; set; }
    public ChatChannel Channel { get; set; } = null!;
    public required int ChannelId { get; set; }
    public required bool IsDeleted { get; set; } = false;
}

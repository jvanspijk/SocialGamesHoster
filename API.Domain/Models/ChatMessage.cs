namespace API.Domain.Models;

public class ChatMessage
{
    public Guid Id { get; init; } = Guid.CreateVersion7();

    public required string Content { get; set; }
    public DateTime SentAt { get; init; } = DateTime.UtcNow;

    public Player? Sender { get; set; }
    public required int? SenderId { get; set; }
    public ChatChannel Channel { get; set; } = null!;
    public required int ChannelId { get; set; }
    public required bool IsDeleted { get; set; } = false;
}

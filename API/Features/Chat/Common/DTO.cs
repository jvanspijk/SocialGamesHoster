namespace API.Features.Chat.Common;

public readonly record struct Message(int Id, string Content, DateTime SentAt, int? SenderId, string? SenderName);


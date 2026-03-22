namespace API.Features.Chat.Common;

public readonly record struct Message(Guid Id, string Content, DateTime SentAt, int? SenderId, string? SenderName);


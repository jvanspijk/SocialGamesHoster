using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessagesFromChannel
{
    public record Request(int ChannelId, DateTime? Before = null, DateTime? After = null, int Limit = 50);

    public record Response(Guid Id, string Content, DateTime SentAt, int? SenderId, string? SenderName)
        : IProjectable<ChatMessage, Response>
    {
        public static Expression<Func<ChatMessage, Response>> Projection =>
            message => new Response(
                message.Id,
                message.Content,
                message.SentAt,
                message.SenderId,
                message.Sender == null ? null : message.Sender.Name);
    }

    public static async Task<IResult> HandleAsync(ChatRepository repo, [AsParameters] Request req)
    {
        var result = await repo.GetFromChannelAsync<Response>(req.ChannelId, req.Before, req.After, req.Limit);
        return result.AsIResult();
    }
}

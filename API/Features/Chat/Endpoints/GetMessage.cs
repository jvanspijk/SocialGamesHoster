using API.DataAccess;
using API.Domain.Entities;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessage
{
    private static string CacheKey(int id) => $"{nameof(GetMessage)}_{id}";
    public record Response(int Id, string Content, DateTime SentAt, string SenderName, int? SenderId, bool IsDeleted) : IProjectable<ChatMessage, Response>
    {
        public static Expression<Func<ChatMessage, Response>> Projection =>
            message => new Response(
                message.Id,
                message.IsDeleted ? "Deleted" : message.Content,
                message.SentAt,
                message.Sender != null ? message.Sender.Name : "Deleted",
                message.SenderId,
                message.IsDeleted
            );
    }
    public static async Task<IResult> HandleAsync(Repository<ChatMessage> repository, IMemoryCache cache, int id)
    {
        Response? message = await cache.GetOrCreateAsync(CacheKey(id), async entry =>
        {
            entry.SlidingExpiration = TimeSpan.FromMinutes(10);
            return await repository.GetReadOnlyAsync<Response>(m => m.Id == id);
        });

        if (message is null)
        {
            return Results.NotFound();
        }           

        return Results.Ok(message);
    }
}

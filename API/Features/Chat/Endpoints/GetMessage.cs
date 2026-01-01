using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessage
{
    public record Response(Guid Id, string Content, DateTime SentAt, string SenderName, int? SenderId, bool IsDeleted) : IProjectable<ChatMessage, Response>
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
    public static async Task<IResult> HandleAsync(ChatRepository repository, IMemoryCache cache, Guid id)
    {
        //TODO: add caching
        // Cache will get invalidated on message delete or player update / delete
        var message = await repository.GetAsync(id);
        if (message is null)
        {
            return Results.NotFound();
        }
        
        return Results.Ok();
    }
}

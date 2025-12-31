using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessage
{
    public record Response(Guid Id, string Content, DateTime SentAt, string PlayerName, bool IsDeleted) : IProjectable<ChatMessage, Response>
    {
        public static Expression<Func<ChatMessage, Response>> Projection =>
            message => new Response(
                message.Id,
                message.IsDeleted ? "Deleted" : message.Content,
                message.SentAt,
                message.Sender != null ? message.Sender.Name : "Deleted",
                message.IsDeleted
            );
    }
    public static async Task<IResult> HandleAsync(ChatRepository repository, Guid id)
    {
        var message = await repository.GetAsync(id);
        if (message is null)
        {
            return Results.NotFound();
        }
        
        return Results.Ok();
    }
}

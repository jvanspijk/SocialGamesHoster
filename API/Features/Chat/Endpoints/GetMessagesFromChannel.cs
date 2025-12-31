using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using System.ComponentModel.DataAnnotations;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessagesFromChannel
{
    public record Request(int ChannelId, DateTime? Before, DateTime? After, int? Limit) : IValidatable<Request>
    {
        public IEnumerable<ValidationError> Validate()
        {
            var errors = new List<ValidationError>();
            if (Limit is not null && Limit <= 0)
            {
                errors.Add(new ValidationError(nameof(Limit), "Limit cannot be negative or zero."));
            }
            return errors;
        }
    }

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
        const int defaultLimit = 50;
        if (req.Limit is null)
        {
            req = req with { Limit = defaultLimit };
        }

        var errors = req.Validate();
        if (errors.Any())
        {
            return Results.ValidationProblem(errors.ToProblemDetails());
        }

        var result = await repo.GetFromChannelAsync<Response>(req.ChannelId, req.Before, req.After, req.Limit.Value);
        return result.AsIResult();
    }
}

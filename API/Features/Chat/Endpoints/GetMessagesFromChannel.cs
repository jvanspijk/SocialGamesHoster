using API.DataAccess;
using API.Domain;
using API.Domain.Entities;
using API.Domain.Validation;
using System.ComponentModel.DataAnnotations;
using System.Diagnostics;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessagesFromChannel
{
    public readonly record struct Request(int ChannelId, DateTime? Before, DateTime? After, int? Limit) : IValidatable<Request>
    {
        public IEnumerable<ValidationError> Validate()
        {
            if (Limit is not null && Limit <= 0)
            {
                yield return new ValidationError(nameof(Limit), "Limit cannot be negative or zero.");
            }
            if(ChannelId <= 0)
            {
                yield return new ValidationError(nameof(ChannelId), "ChannelId must be a positive integer.");
            }
            if (DateTime.UtcNow < After)
            {
                yield return new ValidationError(nameof(After), "After cannot be in the future.");
            }
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

    public static async Task<IResult> HandleAsync(Repository<ChatChannel> repository, [AsParameters] Request req)
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

        var channelExists = await repository.ExistsAsync(req.ChannelId);
        if (!channelExists)
        {
            return Results.NotFound($"Channel with id {req.ChannelId} not found.");
        }

        var result = await repository.GetChannelMessagesAsync<Response>(req.ChannelId, req.Limit.Value, req.Before, req.After);
        return Results.Ok(result);
    }
}

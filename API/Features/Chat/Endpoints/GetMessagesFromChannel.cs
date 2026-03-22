using API.DataAccess;
using API.Domain.Entities;
using System.ComponentModel.DataAnnotations;
using System.Linq.Expressions;

namespace API.Features.Chat.Endpoints;

public static class GetMessagesFromChannel
{
    public readonly record struct Request(int ChannelId, DateTime? Before, DateTime? After, int? MaxMessages) : IValidatableObject
    {       
        public int Limit { get; init; } = MaxMessages ?? 50;
        public IEnumerable<ValidationResult> Validate(ValidationContext validationContext)
        {
            if (Limit <= 0)
            {
                yield return new ValidationResult("Limit must be positive.", [nameof(Limit)]);
            }

            if (ChannelId <= 0)
            {
                yield return new ValidationResult("ChannelId must be a positive integer.", [nameof(ChannelId)]);
            }

            if (After > DateTime.UtcNow)
            {
                yield return new ValidationResult("After cannot be in the future.", [nameof(After)]);
            }

            if (Before.HasValue && After.HasValue && Before < After)
            {
                yield return new ValidationResult("Before date must be after the After date.", [nameof(Before), nameof(After)]);
            }
        }
    }

    public record Response(int Id, string Content, DateTime SentAt, int? SenderId, string? SenderName)
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

    public static async Task<IResult> HandleAsync(IRepository<ChatMessage> repository, [AsParameters] Request req)
    {       
        var channelExists = await repository.ExistsAsync(req.ChannelId);
        if (!channelExists)
        {
            return Results.NotFound($"Channel with id {req.ChannelId} not found.");
        }

        var result = await repository.QueryReadOnlyAsync<Response>(query =>
            query.Where(c => c.ChannelId == req.ChannelId)
                 .Where(m => (!req.Before.HasValue || m.SentAt < req.Before) &&
                             (!req.After.HasValue || m.SentAt > req.After))
                 .OrderByDescending(m => m.SentAt)
                 .Take(req.Limit)
        );
      
        return Results.Ok(result);
    }
}

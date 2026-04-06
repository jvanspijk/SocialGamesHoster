using API.DataAccess;
using API.Domain.Entities;
using API.Features.Auth;
using API.Features.Chat.Hubs;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.AspNetCore.SignalR;
using System.ComponentModel.DataAnnotations;

namespace API.Features.Chat.Endpoints;

public static class SendMessage
{
    public readonly record struct Request(bool IsAdmin, int? PlayerId, string Message) : IValidatableObject
    {
        public IEnumerable<ValidationResult> Validate(ValidationContext validationContext)
        {
            if (string.IsNullOrWhiteSpace(Message))
            {
                yield return new ValidationResult("Message cannot be empty.", [nameof(Message)]);
            }

            if (!IsAdmin && PlayerId is null)
            {
                yield return new ValidationResult("PlayerId is required when IsAdmin is false.", [nameof(PlayerId)]);
            }

            if (IsAdmin && PlayerId is not null)
            {
                yield return new ValidationResult("PlayerId must be null when IsAdmin is true.", [nameof(PlayerId)]);
            }
        }
    }

    public readonly record struct Response(int MessageId, int? PlayerId, int ChannelId, string Message);
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(
        IRepository<ChatMessage> messageRepository, IRepository<ChatChannel> channelRepository,
        AuthService authService, IHubContext<ChatHub, IChatHub> hub, int channelId, Request request,
        HttpRequest httpRequest)
    {
        bool channelExists = await channelRepository.ExistsAsync(channelId);
        if(!channelExists)
        {
            return APIResults.NotFound($"Channel with id {channelId} does not exist.");
        }

        if (request.IsAdmin)
        {
            var adminCheck = authService.IsAdmin(httpRequest);
            if (adminCheck.IsFailure)
            {
                return APIResults.Unauthorized();
            }

            if (!adminCheck.Value)
            {
                return APIResults.Unauthorized();
            }
        }
       
        ChatMessage message = new()
        {
            SenderId = request.PlayerId,
            IsAdmin = request.IsAdmin,
            ChannelId = channelId,
            Content = request.Message,
            IsDeleted = false
        };

        messageRepository.Add(message);
        await messageRepository.SaveChangesAsync();

        var response = new Response(message.Id, message.SenderId, message.ChannelId, message.Content);
        await ChatHub.NotifyMessageSent(hub, channelId, request.PlayerId, message.Id);

        return APIResults.CreatedAtRoute(response, nameof(GetMessage), message.Id);
    }
}

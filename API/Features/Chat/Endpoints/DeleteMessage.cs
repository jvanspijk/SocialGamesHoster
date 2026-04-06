using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Chat.Endpoints;

public static class DeleteMessage
{
    // TODO: add to endpoints file
    public static async Task<Results<Ok, ProblemHttpResult>> HandleAsync(int id, IRepository<ChatMessage> repository)
    {
        var message = await repository.GetWithTrackingAsync(id);
        if (message is null)
        {
            return APIResults.NotFound<ChatMessage>(id);
        }
        message.IsDeleted = true;
        await repository.SaveChangesAsync();
        return APIResults.Ok();
    }
}

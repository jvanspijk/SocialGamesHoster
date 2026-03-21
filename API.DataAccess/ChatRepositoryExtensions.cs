using Microsoft.EntityFrameworkCore;
using API.Domain.Entities;

namespace API.DataAccess;

public static class ChatRepositoryExtensions
{   
    extension(Repository<ChatChannel> repository)
    {
        public async Task<ChatMessage?> GetMessageAsync(Guid id)           
        {
            return await repository.context.Set<ChatMessage>()
                .FindAsync(id);
        }

        public async Task<TResponse[]> GetChannelMessagesAsync<TResponse>(
        int channelId,
        int limit,
        DateTime? before = null,
        DateTime? after = null)
        where TResponse : class, IProjectable<ChatMessage, TResponse>
        {
            var query = repository.context.Set<ChatMessage>()
                .AsNoTracking()
                .Where(m => m.ChannelId == channelId);

            if (before.HasValue)
            {
                query = query.Where(m => m.SentAt < before.Value);
            }

            if (after.HasValue)
            {
                query = query.Where(m => m.SentAt > after.Value);
            }

            return await query.OrderByDescending(m => m.SentAt)
                        .Take(limit)
                        .Select(TResponse.Projection)
                        .ToArrayAsync();
        }

        public ChatMessage AddMessage(ChatMessage message)
        {
            var dbSet = repository.context.Set<ChatMessage>();
            dbSet.Add(message);
            return message;
        }
    }    
}

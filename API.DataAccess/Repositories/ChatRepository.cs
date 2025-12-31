using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class ChatRepository(APIDatabaseContext context)
{
    private readonly APIDatabaseContext _context = context;
    public async Task<Result<ChatMessage>> CreateAsync(int playerId, int channelId, string message)
    {
        bool playerExists = await _context.Players.AnyAsync(p => p.Id == playerId);
        bool channelExists = await _context.ChatChannels.AnyAsync(c => c.Id == channelId);

        if (!playerExists)
        {
            return Errors.ResourceNotFound("Player", playerId);
        }

        if (!channelExists)
        {
            return Errors.ResourceNotFound("ChatChannel", channelId);
        }

        ChatMessage chatMessage = new()
        {
            ChannelId = channelId,
            SenderId = playerId,
            Content = message,
            IsDeleted = false,
        };
        _context.ChatMessages.Add(chatMessage);
        await _context.SaveChangesAsync();
        return chatMessage;
    }

    public async Task<ChatChannel> CreateChannelAsync(string name, int gameId)
    {
        ChatChannel channel = new()
        {
            Name = name,
            GameId = gameId
        };
        _context.ChatChannels.Add(channel);
        await _context.SaveChangesAsync();
        return channel;
    }

    public async Task DeleteAsync(Guid messageId)
    {
        var message = await _context.ChatMessages.FindAsync(messageId);
        if (message != null)
        {
            message.IsDeleted = true;
            await _context.SaveChangesAsync();
        }
    }

    public async Task<Result<List<TProjectable>>> GetFromChannelAsync<TProjectable>(int channelId, DateTime? before, DateTime? after, int limit)
        where TProjectable : class, IProjectable<ChatMessage, TProjectable>
    {
        bool channelExists = await _context.ChatChannels.AnyAsync(c => c.Id == channelId);
        if (!channelExists)
        {
            return Errors.ResourceNotFound("ChatChannel", channelId);
        }            

        var query = _context.ChatMessages
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

        var results = await query
            .OrderByDescending(m => m.SentAt)
            .Take(limit)
            .Select(TProjectable.Projection)
            .ToListAsync();

        return results;
    }    

    public async Task<Result<List<ChatChannel>>> GetChannelsFromGameAsync(int gameId)
    {
        return await _context.ChatChannels
            .AsNoTracking()
            .Where(c => c.GameId == gameId)
            .ToListAsync();
    }

    public Task<ChatMessage?> GetAsync(Guid id)
    {
        throw new NotImplementedException();
    }

    public Task<List<TProjectable>> GetMultipleAsync<TProjectable>(int channelId)
    {
        throw new NotImplementedException();
    }  

    Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : class
    {
        throw new NotImplementedException();
    }

    public async Task<TProjectable?> GetAsync<TProjectable>(Guid id) where TProjectable : class, IProjectable<ChatMessage, TProjectable>
    {
        return await _context.ChatMessages
            .AsNoTracking()
            .Where(m => m.Id == id)
            .Select(TProjectable.Projection)
            .SingleOrDefaultAsync();
    }

    Task<Result<List<TProjectable>>> GetMultipleAsync<TProjectable>(List<int> ids) where TProjectable : class
    {
        throw new NotImplementedException();
    }

   
}

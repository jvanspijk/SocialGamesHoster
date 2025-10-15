using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class RoundRepository : IRepository<Round>
{
    private APIDatabaseContext _context;
    public RoundRepository(APIDatabaseContext context)
    {
        _context = context;
    }
    public IQueryable<Round> AsQueryable()
    {
        return _context.Rounds;
    }

    public Task<Round> CreateAsync(Round entity)
    {
        throw new NotImplementedException();
    }

    public Task DeleteAsync(Round entity)
    {
        throw new NotImplementedException();
    }

    public Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : class, IProjectable<Round, TProjectable>
    {
        throw new NotImplementedException();
    }

    public Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : class, IProjectable<Round, TProjectable>
    {
        throw new NotImplementedException();
    }

    public async Task<Round?> GetAsync(int id)
    {
        return await _context.Rounds.FindAsync(id);
    }

    public async Task<Round?> GetCurrentRoundFromGame(int gameId)
    {
        return await _context.GameSessions
            .Where(g => g.Id == gameId)
            .Select(g => g.CurrentRound)
            .FirstOrDefaultAsync();
    }

    public async Task<Round?> StartNewRound(int gameId, TimeSpan duration)
    {
        var gameSession = await _context.GameSessions.FindAsync(gameId);
        if (gameSession == null)
        {
            return null;
        }

        Round newRound = new(DateTime.UtcNow, duration) { GameId = gameId };

        gameSession.StartNewRound(newRound);
        _context.Rounds.Add(newRound);
        _context.SaveChanges();
        return newRound;
    }

    public Task<Round> UpdateAsync(Round updatedEntity)
    {
        throw new NotImplementedException();
    }
}

using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
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

    public async Task<Result<List<Round>>> GetMultipleAsync(List<int> ids)
    {
        var foundRounds = await _context.Rounds
            .Where(r => ids.Contains(r.Id))
            .ToListAsync();

        if (ids.Count != foundRounds.Count)
        {
            var foundIds = foundRounds.Select(r => r.Id).ToHashSet();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Round", "Ids", string.Join(", ", missingIds));
        }

        return foundRounds;
    }


    public async Task<Round?> GetCurrentRoundFromGame(int gameId)
    {
        return await _context.GameSessions
            .Where(g => g.Id == gameId)
            .Select(g => g.CurrentRound)
            .FirstOrDefaultAsync();
    }

    public async Task<Round?> StartNewRound(int gameId)
    {
        GameSession gameSession = await _context.GameSessions
            .Include(g => g.CurrentRound)
            .Where(g => g.Id == gameId)
            .SingleAsync();

        Round newRound = new() 
        { 
            GameId = gameId, 
            StartedTime = DateTimeOffset.UtcNow 
        };

        gameSession.CurrentRound = newRound;

        _context.Rounds.Add(newRound);

        await _context.SaveChangesAsync();

        return newRound;
    }

    public Task<Round> UpdateAsync(Round updatedEntity)
    {
        throw new NotImplementedException();
    }

    public async Task<Result<bool>> FinishRoundAsync(int roundId)
    {
        var round = await _context.Rounds.FindAsync(roundId);
        if (round == null)
        {
            return Errors.ResourceNotFound(nameof(round), roundId);
        }
        round.FinishedTime = DateTimeOffset.UtcNow;
        await _context.SaveChangesAsync();
        return true;
    }
}

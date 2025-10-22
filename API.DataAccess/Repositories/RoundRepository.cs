using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class RoundRepository(APIDatabaseContext context) : IRepository<Round>
{
    private readonly APIDatabaseContext _context = context;

    #region Create
    public Task<Round> CreateAsync(Round entity)
    {
        //return Errors.InvalidOperation("Use StartNewRound method to create a new round associated with a game session.");
        throw new NotImplementedException();
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
    #endregion

    #region Read

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

    public async Task<Result<Round>> GetCurrentRoundFromGame(int gameId)
    {
        var game = await _context.GameSessions
            .Where(g => g.Id == gameId)
            .Include(g => g.CurrentRound)
            .FirstOrDefaultAsync();

        if (game == null) 
        {
            return Errors.ResourceNotFound(nameof(game), gameId);
        }

        if (game.CurrentRound == null)
        {
            return Errors.ResourceNotFound("Current round for game", gameId);
        }

        return game.CurrentRound;
    }
    #endregion

    #region Update
    public async Task<Round> UpdateAsync(Round updatedRound)
    {
        _context.Entry(updatedRound).State = EntityState.Modified;
        _context.Rounds.Update(updatedRound);
        await _context.SaveChangesAsync();
        return updatedRound;
    }

    public async Task<Result> FinishRoundAsync(int roundId)
    {
        var round = await _context.Rounds.FindAsync(roundId);
        if (round == null)
        {
            return Errors.ResourceNotFound(nameof(round), roundId);
        }
        round.FinishedTime = DateTimeOffset.UtcNow;
        await _context.SaveChangesAsync();
        return Result.Success();
    }

    public async Task<Result> CancelRoundAsync(int roundId)
    {
        var round = await _context.Rounds
            .Include(r => r.GameSession)
            .SingleOrDefaultAsync(r => r.Id == roundId);

        if (round == null)
        {
            return Errors.ResourceNotFound(nameof(round), roundId);
        }

        if(round.GameSession == null)
        {
            return Errors.InvalidOperation("Round is not associated with a game session.");
        }

        if(round.GameSession.CurrentRoundId != round.Id)
        {
            return Errors.InvalidOperation("Round is not the current round of its associated game session.");
        }

        round.GameSession.CurrentRound = null;

        await _context.SaveChangesAsync();
        return Result.Success();
    }
    #endregion

    #region Delete
    public Task DeleteAsync(Round entity)
    {
        throw new NotImplementedException();
    }
    #endregion
}



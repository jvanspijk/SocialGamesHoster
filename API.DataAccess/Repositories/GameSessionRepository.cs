using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class GameSessionRepository(APIDatabaseContext context) : IRepository<GameSession>
{
    private readonly APIDatabaseContext _context = context;

    #region Create
    public async Task<GameSession> CreateAsync(GameSession newGameSession)
    {
        _context.GameSessions.Add(newGameSession);
        await _context.SaveChangesAsync();
        return newGameSession;
    }
    #endregion

    #region Read
    public async Task<TProjectable?> GetAsync<TProjectable>(int id)
       where TProjectable : class, IProjectable<GameSession, TProjectable>
    {
        return await _context.GameSessions
            .Where(a => a.Id == id)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjectable>> GetAllAsync<TProjectable>()
        where TProjectable : class, IProjectable<GameSession, TProjectable>
    {
        return await _context.GameSessions
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<Result<List<TProjectable>>> GetMultipleAsync<TProjectable>(List<int> ids)
    where TProjectable : class, IProjectable<GameSession, TProjectable>
    {
        var foundRulesets = await _context.GameSessions
            .Where(r => ids.Contains(r.Id))
            .Select(TProjectable.Projection)
            .ToListAsync();

        if (ids.Count != foundRulesets.Count)
        {
            var foundIds = await _context.GameSessions
                .Where(r => ids.Contains(r.Id))
                .Select(r => r.Id)
                .ToHashSetAsync();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Game sessions", "Ids", string.Join(", ", missingIds));
        }

        return foundRulesets;
    }

    public async Task<List<TProjectable>> GetAllActiveAsync<TProjectable>()
        where TProjectable : class, IProjectable<GameSession, TProjectable>
    {

        return await _activeSessions
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<int> GetActiveSessionId()
    {
        return await _activeSessions
            .Select(gs => gs.Id)
            .SingleOrDefaultAsync();
    }

    public Task<GameSession?> GetAsync(int id)
    {
        return _context.GameSessions
            .Include(gs => gs.Participants)
            .Include(gs => gs.Winners)
            .Include(gs => gs.CurrentRound)
            .FirstOrDefaultAsync(gs => gs.Id == id);
    }

    public async Task<Result<List<GameSession>>> GetMultipleAsync(List<int> ids)
    {
        var foundSessions = await _context.GameSessions
            .Include(gs => gs.Winners)
            .Where(gs => ids.Contains(gs.Id))
            .ToListAsync();

        if(ids.Count != foundSessions.Count)
        {
            var foundIds = foundSessions.Select(gs => gs.Id).ToHashSet();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("GameSession", "Ids", string.Join(", ", missingIds));
        }

        return foundSessions;
    }
    #endregion

    #region Update
    public async Task<GameSession> UpdateAsync(GameSession updatedGame)
    {
        _context.Entry(updatedGame).State = EntityState.Modified;
        _context.GameSessions.Update(updatedGame);
        await _context.SaveChangesAsync();
        return updatedGame;
    }   

    public async Task<Result<GameSession>> StartGameSession(int gameSessionId)
    {
        GameSession? session = await _context.GameSessions
            .FindAsync(gameSessionId);

        if (session == null)
        {
            return Errors.ResourceNotFound(nameof(session), gameSessionId);
        }

        switch (session.Status)
        {
            case GameStatus.Running:
            case GameStatus.Paused:
                return Errors.InvalidOperation($"Game session with id {gameSessionId} is already running.");
            case GameStatus.Finished:
                return Errors.InvalidOperation($"Game session with id {gameSessionId} has already finished.");
        }

        session.Status = GameStatus.Running;
        await _context.SaveChangesAsync();
        return session;
    }

    public async Task<Result<GameSession>> FinishGameSession(int gameSessionId)
    {
        GameSession? session = await _context.GameSessions
            .Include(gs => gs.CurrentRound)
            .FirstOrDefaultAsync(gs => gs.Id == gameSessionId);

        if (session == null)
        {
            return Errors.ResourceNotFound(nameof(session), gameSessionId);
        }

        if(session.IsDone)
        {
            return Errors.InvalidOperation($"Game session with id {gameSessionId} has already completed.");
        }

        if(session.CurrentRound != null && !session.CurrentRound.FinishedTime.HasValue)
        {
            session.CurrentRound.FinishedTime = DateTimeOffset.UtcNow;
        }

        session.Status = GameStatus.Finished;
        await _context.SaveChangesAsync();
        return session;
    }

    public async Task<Result<GameSession>> CancelGameSession(int gameSessionId)
    {
        GameSession? session = await _context.GameSessions
            .FirstOrDefaultAsync(gs => gs.Id == gameSessionId);

        if (session == null)
        {
            return Errors.ResourceNotFound(nameof(session), gameSessionId);
        }

        if (session.IsDone)
        {
            return Errors.InvalidOperation($"Game session with id {gameSessionId} has already completed.");
        }

        session.Status = GameStatus.Cancelled;
        await _context.SaveChangesAsync();
        return session;
    }
    #endregion

    #region Delete
    public async Task<Result> DeleteAsync(int id)
    {
        if(!_context.Database.IsRelational())
        {
            var session = await _context.GameSessions.FindAsync(id);
            if (session == null) return Errors.ResourceNotFound("Session", id);
            _context.GameSessions.Remove(session);
            await _context.SaveChangesAsync();
            return Result.Success();
        }

        int rowsDeleted = await _context.GameSessions
            .Where(g => g.Id == id)
            .ExecuteDeleteAsync();

        if (rowsDeleted == 0)
        {
            return Errors.ResourceNotFound("Game session", id);
        }

        return Result.Success();
    }
    public Task DeleteAsync(GameSession entity)
    {
        throw new NotImplementedException();
    }
    #endregion

    private IQueryable<GameSession> _activeSessions => _context.GameSessions
        .Where(gs => gs.Status == GameStatus.Running || gs.Status == GameStatus.Paused);

    private IQueryable<GameSession> _finishedSessions => _context.GameSessions
        .Where(gs => gs.Status == GameStatus.Finished);
}

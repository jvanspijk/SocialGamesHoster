using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class GameSessionRepository(APIDatabaseContext context) : IRepository<GameSession>
{
    private readonly APIDatabaseContext _context = context;

    public IQueryable<GameSession> AsQueryable()
    {
        return _context.GameSessions.AsQueryable();
    }

    public Task<GameSession> CreateAsync(GameSession entity)
    {
        throw new NotImplementedException();
    }

    public Task DeleteAsync(GameSession entity)
    {
        throw new NotImplementedException();
    }

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

    public async Task<List<TProjectable>> GetAllActiveAsync<TProjectable>()
        where TProjectable : class, IProjectable<GameSession, TProjectable>
    {

        return await GetActiveSessionsAsQueryable()
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<int> GetActiveSessionId()
    {
        return await GetActiveSessionsAsQueryable()
            .Select(gs => gs.Id)
            .SingleOrDefaultAsync();
    }

    public Task<GameSession?> GetAsync(int id)
    {
        return _context.GameSessions
            .Include(gs => gs.Ruleset!.Name)
            .Include(gs => gs.Participants)
            .Include(gs => gs.Winners)
            .FirstOrDefaultAsync(gs => gs.Id == id);
    }

    public Task<GameSession> UpdateAsync(GameSession updatedEntity)
    {
        throw new NotImplementedException();
    }

    private IQueryable<GameSession> GetActiveSessionsAsQueryable() {
        return _context.GameSessions
        .Where(gs => gs.Status == GameStatus.Running || gs.Status == GameStatus.Paused);
    }

    public async Task<Result<GameSession>> StartGameSession(int gameSessionId)
    {
        GameSession? session = await _context.GameSessions
            .FirstOrDefaultAsync(gs => gs.Id == gameSessionId);

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

    public async Task<Result<GameSession>> EndGameSession(int gameSessionId)
    {
        GameSession? session = await _context.GameSessions
            .Include(gs => gs.CurrentRound)
            .FirstOrDefaultAsync(gs => gs.Id == gameSessionId);

        if (session == null)
        {
            return Errors.ResourceNotFound(nameof(session), gameSessionId);
        }
        if(session.Status == GameStatus.Finished)
        {
            return Errors.InvalidOperation($"Game session with id {gameSessionId} has already finished.");
        }

        if(session.CurrentRound != null && !session.CurrentRound.FinishedTime.HasValue)
        {
            session.CurrentRound.FinishedTime = DateTimeOffset.UtcNow;
        }

        session.Status = GameStatus.Finished;
        await _context.SaveChangesAsync();
        return session;
    }
}

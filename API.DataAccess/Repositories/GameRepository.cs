using API.Domain.Models;
using Microsoft.EntityFrameworkCore;
using System.Data;

namespace API.DataAccess.Repositories;

public class GameRepository : IRepository<Game>
{
    private readonly APIDatabaseContext _context;
    public GameRepository(APIDatabaseContext context)
    {
        _context = context;
    }
    public IQueryable<Game> AsQueryable() => _context.Games.AsNoTracking();

    // Create
    public async Task<Game> CreateAsync(Game game)
    {
        _context.Games.Add(game);
        await _context.SaveChangesAsync();
        return game;
    }

    // Read
    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : IProjectable<Game, TProjectable>
    {
        return await _context.Games
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : IProjectable<Game, TProjectable>
    {
        return await _context.Games
            .Where(a => a.Id == id)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }

    // Update
    public async Task UpdateAsync(Game game)
    {
        _context.Entry(game).State = EntityState.Modified;
        _context.Games.Update(game);
        await _context.SaveChangesAsync();
    }

    // Delete
    public async Task DeleteAsync(Game game)
    {
        _context.Entry(game).State = EntityState.Modified;
        _context.Games.Remove(game);
        await _context.SaveChangesAsync();
    }
}

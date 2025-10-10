using API.Domain.Models;
using Microsoft.EntityFrameworkCore;
using System.Data;

namespace API.DataAccess.Repositories;

public class GameRepository : IRepository<Ruleset>
{
    private readonly APIDatabaseContext _context;
    public GameRepository(APIDatabaseContext context)
    {
        _context = context;
    }
    public IQueryable<Ruleset> AsQueryable() => _context.Rulesets.AsNoTracking();

    // Create
    public async Task<Ruleset> CreateAsync(Ruleset game)
    {
        _context.Rulesets.Add(game);
        await _context.SaveChangesAsync();
        return game;
    }

    // Read
    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : IProjectable<Ruleset, TProjectable>
    {
        return await _context.Rulesets
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : IProjectable<Ruleset, TProjectable>
    {
        return await _context.Rulesets
            .Where(a => a.Id == id)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<Ruleset?> GetAsync(int id)
    {
        return await _context.Rulesets.FindAsync(id);
    }

    // Update
    public async Task UpdateAsync(Ruleset updatedGame)
    {
        _context.Entry(updatedGame).State = EntityState.Modified;
        _context.Rulesets.Update(updatedGame);
        await _context.SaveChangesAsync();
    }

    // Delete
    public async Task DeleteAsync(Ruleset game)
    {
        _context.Entry(game).State = EntityState.Modified;
        _context.Rulesets.Remove(game);
        await _context.SaveChangesAsync();
    }
}

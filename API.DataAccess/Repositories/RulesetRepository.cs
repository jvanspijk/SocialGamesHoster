using API.Domain.Models;
using Microsoft.EntityFrameworkCore;
using System.Data;

namespace API.DataAccess.Repositories;

public class RulesetRepository : IRepository<Ruleset>
{
    private readonly APIDatabaseContext _context;
    public RulesetRepository(APIDatabaseContext context)
    {
        _context = context;
    }
    public IQueryable<Ruleset> AsQueryable() => _context.Rulesets.AsNoTracking();

    // Create
    public async Task<Ruleset> CreateAsync(Ruleset ruleset)
    {
        _context.Rulesets.Add(ruleset);
        await _context.SaveChangesAsync();
        return ruleset;
    }

    // Read
    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : 
        class, IProjectable<Ruleset, TProjectable>
    {
        return await _context.Rulesets
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : 
        class, IProjectable<Ruleset, TProjectable>
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
    public async Task<Ruleset> UpdateAsync(Ruleset updatedRuleset)
    {
        _context.Entry(updatedRuleset).State = EntityState.Modified;
        _context.Rulesets.Update(updatedRuleset);
        await _context.SaveChangesAsync();
        return updatedRuleset;
    }

    // Delete
    public async Task DeleteAsync(Ruleset ruleset)
    {
        _context.Entry(ruleset).State = EntityState.Modified;
        _context.Rulesets.Remove(ruleset);
        await _context.SaveChangesAsync();
    }
}

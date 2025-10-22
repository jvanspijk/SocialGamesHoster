using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;
using System.Data;

namespace API.DataAccess.Repositories;

public class RulesetRepository(APIDatabaseContext context) : IRepository<Ruleset>
{
    private readonly APIDatabaseContext _context = context;

    #region Create
    public async Task<Ruleset> CreateAsync(Ruleset ruleset)
    {
        _context.Rulesets.Add(ruleset);
        await _context.SaveChangesAsync();
        return ruleset;
    }
    #endregion

    #region Read
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

    public async Task<Result<List<TProjectable>>> GetMultipleAsync<TProjectable>(List<int> ids) 
        where TProjectable : class, IProjectable<Ruleset, TProjectable>
    {
        var foundRulesets = await _context.Rulesets
            .Where(r => ids.Contains(r.Id))
            .Select(TProjectable.Projection)
            .ToListAsync();

        if (ids.Count != foundRulesets.Count)
        {
            var foundIds = await _context.Rulesets
                .Where(r => ids.Contains(r.Id))
                .Select(r => r.Id)
                .ToHashSetAsync();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Rulesets", "Ids", string.Join(", ", missingIds));
        }

        return foundRulesets;
    }

    public async Task<Ruleset?> GetAsync(int id)
    {
        return await _context.Rulesets.FindAsync(id);
    }

    public async Task<Result<List<Ruleset>>> GetMultipleAsync(List<int> ids)
    {
        var foundRulesets = await _context.Rulesets
            .Where(r => ids.Contains(r.Id))
            .ToListAsync();

        if(ids.Count != foundRulesets.Count)
        {
            var foundIds = foundRulesets.Select(r => r.Id).ToHashSet();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Ruleset", "Ids", string.Join(", ", missingIds));
        }

        return foundRulesets;
    }
    #endregion

    #region Update
    public async Task<Ruleset> UpdateAsync(Ruleset updatedRuleset)
    {
        _context.Entry(updatedRuleset).State = EntityState.Modified;
        _context.Rulesets.Update(updatedRuleset);
        await _context.SaveChangesAsync();
        return updatedRuleset;
    }
    #endregion

    #region Delete
    public async Task DeleteAsync(Ruleset ruleset)
    {
        _context.Entry(ruleset).State = EntityState.Modified;
        _context.Rulesets.Remove(ruleset);
        await _context.SaveChangesAsync();
    }
    #endregion
}

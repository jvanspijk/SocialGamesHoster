using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;
namespace API.DataAccess.Repositories;

public class AbilityRepository : IRepository<Ability>
{    
    private readonly APIDatabaseContext _context;
    public AbilityRepository(APIDatabaseContext context)
    {
        _context = context;
    }

    public IQueryable<Ability> AsQueryable() => _context.Abilities.AsNoTracking();

    // Read
    public async Task<TProjectable?> GetAsync<TProjectable>(int id)
        where TProjectable : class, IProjectable<Ability, TProjectable>
    {
        return await _context.Abilities
            .Where(a => a.Id == id)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() 
        where TProjectable : class, IProjectable<Ability, TProjectable>
    {
        return await _context.Abilities
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<List<TProjectable>> GetAllFromRulesetAsync<TProjectable>(int rulesetId) 
        where TProjectable : class, IProjectable<Ability, TProjectable>
    {
        return await _context.Abilities
            .Where(a => a.RulesetId == rulesetId)
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<Ability?> GetAsync(int id)
    {
        return await _context.Abilities.FindAsync(id);
    }

    public async Task<Result<List<Ability>>> GetAsync(List<int> abilityIds)
    {
        var abilities = await _context.Abilities.Where(a => abilityIds.Contains(a.Id)).ToListAsync();
        if(abilityIds.Count != abilities.Count)
        {
            var foundIds = abilities.Select(a => a.Id).ToHashSet();
            var missingIds = abilityIds.Where(id => !foundIds.Contains(id));
            Errors.ResourceNotFound("Ability", "Ids", string.Join(", ", missingIds));
        }
        return abilities;
    }

    public async Task<Ability> CreateAsync(Ability ability)
    {
        _context.Abilities.Add(ability);
        await _context.SaveChangesAsync();
        return ability;
    }

    public async Task<List<Ability>> GetAllAsync()
    {
        return await _context.Abilities.ToListAsync();        
    }    

    public async Task<Ability> UpdateAsync(Ability updatedAbility)
    { 
        _context.Entry(updatedAbility).State = EntityState.Modified;
        _context.Abilities.Update(updatedAbility);
        await _context.SaveChangesAsync();
        return updatedAbility;
    }

    public async Task DeleteAsync(Ability ability)
    {
        _context.Entry(ability).State = EntityState.Modified;
        _context.Abilities.Remove(ability);
        await _context.SaveChangesAsync();
    }
}

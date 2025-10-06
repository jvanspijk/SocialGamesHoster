using API.Domain.Models;
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

    public async Task<TProjection?> GetAsync<TProjection>(int id) where TProjection : IProjectable<Ability, TProjection>
    {
        return await _context.Abilities
            .Where(a => a.Id == id)
            .Select(TProjection.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjection>> GetAllAsync<TProjection>() where TProjection : IProjectable<Ability, TProjection>
    {
        return await _context.Abilities
            .Select(TProjection.Projection)
            .ToListAsync();
    }

    public async Task<Ability> CreateAsync(Ability ability)
    {
        _context.Abilities.Add(ability);
        await _context.SaveChangesAsync();
        return ability;
    }
    
    public async Task<Ability?> GetByIdAsync(int id)
    {
        return await _context.Abilities.FindAsync(id);        
    }

    public async Task<List<Ability>> GetAllAsync()
    {
        return await _context.Abilities.ToListAsync();        
    }    

    public async Task UpdateAsync(Ability ability)
    {
        _context.Entry(ability).State = EntityState.Modified;
        _context.Abilities.Update(ability);
        await _context.SaveChangesAsync();        
    }

    public async Task DeleteAsync(Ability ability)
    {
        _context.Entry(ability).State = EntityState.Modified;
        _context.Abilities.Remove(ability);
        await _context.SaveChangesAsync();
    }


}

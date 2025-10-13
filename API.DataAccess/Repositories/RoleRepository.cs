using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class RoleRepository : IRepository<Role>
{
    private readonly APIDatabaseContext _context;

    public RoleRepository(APIDatabaseContext context) 
    {
        _context = context;
    }

    public IQueryable<Role> AsQueryable() => _context.Roles.AsNoTracking();

    // Create
    public async Task<Role> CreateAsync(Role role)
    {
        _context.Roles.Add(role);
        await _context.SaveChangesAsync();
        return role;
    }

    // Read
    public async Task<TProjectable?> GetAsync<TProjectable>(int id) where TProjectable : IProjectable<Role, TProjectable>
    {
        return await _context.Roles
            .Where(a => a.Id == id)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() where TProjectable : IProjectable<Role, TProjectable>
    {
        return await _context.Roles
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<List<TProjectable>> GetAllFromRulesetAsync<TProjectable>(int rulesetId) where TProjectable : IProjectable<Role, TProjectable>
    {
        return await _context.Roles
            .Where(r => r.RulesetId == rulesetId)
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<Role?> GetAsync(int id)
    {
        return await _context.Roles.FindAsync(id);
    }

    // Update
    public async Task<Role> UpdateAsync(Role updatedRole)
    {
        _context.Entry(updatedRole).State = EntityState.Modified;
        _context.Roles.Update(updatedRole);
        await _context.SaveChangesAsync();
        return updatedRole;
    }

    // Delete
    public async Task DeleteAsync(Role role)
    {
        _context.Entry(role).State = EntityState.Modified;
        _context.Roles.Remove(role);
        await _context.SaveChangesAsync();
    }


}
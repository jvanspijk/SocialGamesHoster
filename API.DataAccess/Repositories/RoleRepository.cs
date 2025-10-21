using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
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
    public async Task<TProjectable?> GetAsync<TProjectable>(int id)
        where TProjectable : class, IProjectable<Role, TProjectable>
    {
        return await _context.Roles
            .Where(r => r.Id == id)
            .Select(TProjectable.Projection)            
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() 
        where TProjectable : class, IProjectable<Role, TProjectable>
    {
        return await _context.Roles
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<List<TProjectable>> GetAllFromRulesetAsync<TProjectable>(int rulesetId) 
        where TProjectable : class, IProjectable<Role, TProjectable>
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

    public async Task<Result<List<Role>>> GetMultipleAsync(List<int> ids)
    {
        var foundRoles = await _context.Roles
            .Where(r => ids.Contains(r.Id))
            .ToListAsync();

        if(ids.Count != foundRoles.Count)
        {
            var foundIds = foundRoles.Select(r => r.Id).ToHashSet();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Role", "Ids", string.Join(", ", missingIds));
        }

        return foundRoles;
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
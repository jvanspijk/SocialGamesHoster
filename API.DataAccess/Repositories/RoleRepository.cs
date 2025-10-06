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

    public async Task<TProjection?> GetAsync<TProjection>(int id) where TProjection : IProjectable<Role, TProjection>
    {
        return await _context.Roles
            .Where(a => a.Id == id)
            .Select(TProjection.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjection>> GetAllAsync<TProjection>() where TProjection : IProjectable<Role, TProjection>
    {
        return await _context.Roles
            .Select(TProjection.Projection)
            .ToListAsync();
    }

    // Update

    public async Task UpdateAsync(Role role)
    {
        _context.Entry(role).State = EntityState.Modified;
        _context.Roles.Update(role);
        await _context.SaveChangesAsync();
    }

    // Delete
    public async Task DeleteAsync(Role role)
    {
        _context.Entry(role).State = EntityState.Modified;
        _context.Roles.Remove(role);
        await _context.SaveChangesAsync();
    }


}
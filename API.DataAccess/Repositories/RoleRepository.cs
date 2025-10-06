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

    public async Task<List<Role>> GetAllAsync()
    {
        _context.ChangeTracker.QueryTrackingBehavior = QueryTrackingBehavior.NoTracking;

        return await _context.Roles
            .Include(r => r.Abilities)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
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
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
    // Create
    public async Task<Role> CreateAsync(Role role)
    {
        _context.Roles.Add(role);
        await _context.SaveChangesAsync();
        return role;
    }

    // Read

    public async Task<Role?> GetByIdAsync(int roleId)
    {
        return await _context.Roles
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
            .SingleOrDefaultAsync(r => r.Id == roleId);
    }

    public async Task<Role?> GetByNameAsync(string name)
    {
        return await _context.Roles
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
            .SingleOrDefaultAsync(r => r.Name == name);
    }

    public async Task<Role?> GetByPlayerNameAsync(string playerName)
    {
        var playerQuery = _context.Players
            .Where(player => player.Name == playerName && player.Role != null)
            .Include(player => player.Role!)
                .ThenInclude(role => role.AbilityAssociations)
                    .ThenInclude(ra => ra.Ability)
            .Include(player => player.Role!)
                .ThenInclude(role => role.CanSee)
            .Include(player => player.Role!)
                .ThenInclude(role => role.CanBeSeenBy)
            .SingleOrDefaultAsync();

        var player = await playerQuery;
        return player?.Role;
    }

    public async Task<List<Role>> GetAllAsync()
    {
        _context.ChangeTracker.QueryTrackingBehavior = QueryTrackingBehavior.NoTracking;

        return await _context.Roles
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
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
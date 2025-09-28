using API.DTO;
using API.Models;
using API.Validation;
using Microsoft.EntityFrameworkCore;
using System.Xml.Linq;

namespace API.DataAccess.Repositories;

public class RoleRepository
{
    private readonly APIDatabaseContext _context;

    public RoleRepository(APIDatabaseContext context) 
    {
        _context = context;
    }

    public async Task<Result<List<Player>>> GetPlayersWithRoleAsync(int roleId)
    {
        var role = await _context.Roles
            .Include(r => r.PlayersWithRole)
            .SingleOrDefaultAsync(r => r.Id == roleId);

        if(role == null)
        {
            return Errors.ResourceNotFound("Role", roleId.ToString());
        }

        return role.PlayersWithRole.ToList();
    }

    public async Task<Result<Role>> GetFromNameAsync(string name)
    {
        Role? role = await _context.Roles
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
            .SingleOrDefaultAsync(r => r.Name == name);
        if (role == null)
        {
            return Errors.ResourceNotFound($"Role for player {name} not found");
        }
        return role;
    }

    public async Task<Result<Role>> GetFromIdAsync(int id)
    {       
        Role? role = await _context.Roles
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
            .SingleOrDefaultAsync(r => r.Id == id);

        if (role == null)
        {
            return Errors.ResourceNotFound("Role", id.ToString());
        }

        return role;        
    }

    public async Task<Result<ICollection<RoleDTO>>> GetAllAsync()
    {
        _context.ChangeTracker.QueryTrackingBehavior = QueryTrackingBehavior.NoTracking;

        var roles = await _context.Roles
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .ToListAsync();
        var roleDtos = roles.Select(static a => new RoleDTO(a)).ToList();
        return roleDtos;
    }

    // TODO: is this not the same as GetFromName?
    public async Task<Result<Role>> GetRoleByPlayerNameAsync(string playerName)
    {
        Role? role = await _context.Roles
            .Include(r => r.AbilityAssociations).ThenInclude(ra => ra.Ability)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
            .Where(r => r.PlayersWithRole.Any(p => p.Name == playerName))
            .SingleOrDefaultAsync();

        if (role == null)
        {
            return Errors.ResourceNotFound($"Role for player {playerName} not found");
        }

        return role;
    }
}
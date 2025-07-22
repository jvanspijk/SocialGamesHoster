using API.DTO;
using API.Models;
using API.Validation;
using LanguageExt;
using LanguageExt.Common;
using Microsoft.EntityFrameworkCore;

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
        try
        {
            var role = await _context.Roles
                .Include(r => r.PlayersWithRole)
                .SingleOrDefaultAsync(r => r.Id == roleId);

            if (role == null || !role.PlayersWithRole.Any())
            {
                return new Result<List<Player>>(
                    new NotFoundException($"No players found for role with id {roleId}.")
                );
            }

            return role.PlayersWithRole.ToList();
        }
        catch (Exception ex)
        {
            // TODO: Log exception and explicitly handle known ones
            return new Result<List<Player>>(ex);
        }
    }

    public async Task<Result<Role>> GetFromNameAsync(string name)
    {
        try
        {
            Role? role = await _context.Roles
                .Include(r => r.AbilityAssociations)
                    .ThenInclude(ra => ra.Ability)
                .Include(r => r.CanSee)
                .Include(r => r.CanBeSeenBy)
                .SingleOrDefaultAsync(r => r.Name == name);
            if (role == null)
            {
                return new Result<Role>(new NotFoundException($"Role with name '{name}' not found."));
            }
            return role;
        }
        catch (Exception ex)
        {
            // TODO: Log exception and explicitly handle known ones
            return new Result<Role>(ex);
        }
    }

    public async Task<Result<Role>> GetFromIdAsync(int id)
    {
        try
        {
            Role? role = await _context.Roles
                .Include(r => r.AbilityAssociations)
                    .ThenInclude(ra => ra.Ability)
                .Include(r => r.CanSee)
                .Include(r => r.CanBeSeenBy)
                .SingleOrDefaultAsync(r => r.Id == id);

            if (role == null)
            {
                return new Result<Role>(new NotFoundException($"Role with id {id} not found."));
            }

            return role;
        }
        catch (Exception ex)
        {
            // TODO: Log exception and explicitly handle known ones
            return new Result<Role>(ex);
        }
    }

    public async Task<Result<ICollection<RoleDTO>>> GetAllAsync()
    {
        _context.ChangeTracker.QueryTrackingBehavior = QueryTrackingBehavior.NoTracking;
        try
        {
            var roles = await _context.Roles
                .Include(r => r.AbilityAssociations)
                    .ThenInclude(ra => ra.Ability)
                .ToListAsync();
            var roleDtos = roles.Select(RoleDTO.FromModel).ToList();
            return new Result<ICollection<RoleDTO>>(roleDtos);
        }
        catch (Exception ex)
        {
            // TODO: Log exception and explicitly handle known ones
            return new Result<ICollection<RoleDTO>>(ex);
        }
    }

    public async Task<Result<Role>> GetRoleByPlayerNameAsync(string playerName)
    {
        try
        {
            Role? role = await _context.Roles
             .Include(r => r.AbilityAssociations).ThenInclude(ra => ra.Ability)
             .Include(r => r.CanSee)
             .Include(r => r.CanBeSeenBy)
             .Where(r => r.PlayersWithRole.Any(p => p.Name == playerName))
             .SingleOrDefaultAsync();

            if (role == null)
            {
                return new Result<Role>(
                    new NotFoundException($"Role not found for player {playerName}.")
                );
            }

            return role;
        }
        catch (Exception ex)
        {
            // TODO: Log exception and explicitly handle known ones
            return new Result<Role>(ex);
        }
    }
}
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

    public RoleRepository(APIDatabaseContext context) {
        _context = context;
    }

    public async Task<Result<Role>> GetFromIdAsync(int id)
    {        
        Role? role = await _context.Roles.FirstOrDefaultAsync(r => r.Id == id);

        if (role == null)
        {
            var exception = new NotFoundException($"Role with id {id} not found.");
            return new Result<Role>(exception);
        }

        return role;
    }

    public async Task<Result<RoleDTO>> GetFromPlayerAsync(string username)
    {
        try
        {
            Role? role = await _context.Players
                .AsNoTracking()
                .Include(p => p.Role)
                    .ThenInclude(r => r.AbilityAssociations)
                        .ThenInclude(ra => ra.Ability)
                .SingleAsync(p => p.RoleId != null && p.Name == username)
                .Select(p => p.Role);

            if (role == null)
            {
                return new Result<RoleDTO>(new NotFoundException($"Role for player '{username}' not found."));
            }

            RoleDTO roleDto = RoleDTO.FromModel(role);
            if (roleDto == null)
            {
                return new Result<RoleDTO>(new Exception($"Role DTO for player '{username}' could not be created."));
            }

            return new Result<RoleDTO>(roleDto);
        }
        catch (Exception ex)
        {
            return new Result<RoleDTO>(ex);
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
            return new Result<ICollection<RoleDTO>>(ex);
        }
    }
}
using API.Models;
using API.Validation;
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

    public async Task<Result<Role>> GetFromPlayerAsync(string username)
    {
        try
        {
        Role? role = await _context.Players
            .Where(p => p.Name.ToLower() == username.ToLower() && p.RoleId != null)
            .Include(p => p.Role)
                .ThenInclude(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .Select(p => p.Role)
            .FirstOrDefaultAsync();

            if (role != null)
            {
                return new Result<Role>(role);
            }
            else
            {
                return new Result<Role>(new NotFoundException($"Role for player '{username}' not found."));
            }
        }
        catch (Exception ex)
        {
            return new Result<Role>(ex);
        }
    }
}
using API.DTO;
using API.Models;
using API.Validation;
using Microsoft.EntityFrameworkCore;
using Npgsql;
namespace API.DataAccess.Repositories;

public class AbilityRepository
{
    private readonly APIDatabaseContext _context;
    public AbilityRepository(APIDatabaseContext context)
    {
        _context = context;
    }

    public async Task<Result<AbilityDTO>> GetAbilityAsync(int id)
    {
        Ability? ability = await _context.Abilities.FindAsync(id);
        if (ability == null)
        {
            return Errors.ResourceNotFound("Ability", id.ToString());
        }
        return new AbilityDTO(ability);
    }

    public async Task<Result<IEnumerable<AbilityDTO>>> GetAllAbilitiesAsync()
    {
        IEnumerable<Ability> abilities = await _context.Abilities.ToListAsync();
        return abilities.Select(static a => new AbilityDTO(a)).ToList();
    }
}

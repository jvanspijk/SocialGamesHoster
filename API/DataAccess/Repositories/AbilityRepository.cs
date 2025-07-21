using API.DTO;
using API.Models;
using API.Validation;
using LanguageExt.Common;
using Microsoft.EntityFrameworkCore;

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
        try
        {
            Ability? ability = await _context.Abilities.FindAsync(id);
            if (ability == null)
            {
                return new Result<AbilityDTO>(new NotFoundException($"Ability with id {id} not found."));
            }
            return AbilityDTO.FromModel(ability);
        }
        catch (Exception ex)
        {
            return new Result<AbilityDTO>(ex);
        }
    }

    public async Task<Result<IEnumerable<AbilityDTO>>> GetAllAbilitiesAsync()
    {
        try
        {
            IEnumerable<Ability> abilities = await _context.Abilities.ToListAsync();
            return abilities.Select(a => AbilityDTO.FromModel(a)).ToList();
        }
        catch (Exception ex)
        {
            return new Result<IEnumerable<AbilityDTO>>(ex);
        }
    }

}

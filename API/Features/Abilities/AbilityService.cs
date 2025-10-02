using API.DataAccess.Repositories;
using API.Features.Abilities.Responses;
using API.Models;
using API.Validation;

namespace API.Features.Abilities;

public class AbilityService
{
    private readonly AbilityRepository _repository;
    public AbilityService(AbilityRepository abilityRepository)
    {
        _repository = abilityRepository;
    }
    public async Task<Result<AbilityResponse>> GetAsync(int id)
    {
        Ability? ability = await _repository.GetByIdAsync(id);
        if(ability is null)
        {
            return Errors.ResourceNotFound("Ability", id);
        }
        return new AbilityResponse(ability);        
    }

    public async Task<Result<List<AbilityResponse>>> GetAllAsync()
    {
        List<Ability> abilities = await _repository.GetAllAsync();
        return abilities.Select(ab => new AbilityResponse(ab)).ToList();
    }
}

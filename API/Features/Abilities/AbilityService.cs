using API.DataAccess.Repositories;
using API.Features.Abilities.Responses;

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
        var ability = await _repository.GetAbilityAsync(id);
        return ability.Map(a => new AbilityResponse(a));        
    }

    public async Task<Result<IEnumerable<AbilityResponse>>> GetAllAsync()
    {
        var abilities = await _repository.GetAllAsync();
        return abilities.Map(ab => ab.Select(a => new AbilityResponse(a)));
    }
}

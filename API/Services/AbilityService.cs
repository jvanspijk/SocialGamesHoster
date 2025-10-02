using API.DataAccess.Repositories;
using API.Models.Requests;

namespace API.Services;

public class AbilityService
{
    private readonly AbilityRepository _repository;
    public AbilityService(AbilityRepository abilityRepository)
    {
        _repository = abilityRepository;
    }
    public async Task<Result<AbilityDTO>> GetAsync(int id)
    {
        var ability = await _repository.GetAbilityAsync(id);
        return ability.Map(a => new AbilityDTO(a));        
    }

    public async Task<Result<IEnumerable<AbilityDTO>>> GetAllAsync()
    {
        var abilities = await _repository.GetAllAsync();
        return abilities.Map(ab => ab.Select(a => new AbilityDTO(a)));
    }
}

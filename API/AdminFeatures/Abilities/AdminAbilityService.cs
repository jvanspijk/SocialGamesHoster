using API.AdminFeatures.Abilities.Responses;
using API.Domain;

namespace API.AdminFeatures.Abilities;

public class AdminAbilityService
{
    public async Task<Result<AbilityResponse>> GetAsync(int id)
    {
        throw new NotImplementedException();
    }

    public async Task<Result<AbilityResponse>> GetByNameAsync(string name)
    {
        throw new NotImplementedException();
    }

    public async Task<Result<List<AbilityResponse>>> GetAllAsync()
    {
        throw new NotImplementedException();
    }
}

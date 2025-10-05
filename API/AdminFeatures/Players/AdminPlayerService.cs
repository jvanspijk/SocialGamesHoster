using API.Domain;
using API.Features.Players.Responses;

namespace API.AdminFeatures.Players;

public class AdminPlayerService
{
    public async Task<Result<PlayerResponse>> GetAsync(int id)
    {
        throw new NotImplementedException();
    }

    public async Task<Result<PlayerResponse>> GetByNameAsync(string name)
    {
        throw new NotImplementedException();
    }

    public async Task<Result<List<PlayerResponse>>> GetAllAsync()
    {
        throw new NotImplementedException();
    }
}

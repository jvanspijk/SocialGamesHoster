using API.Models;
using API.Validation;

namespace API.DataAccess.Repositories;

public class RoundRepository
{
    private Round _currentRound = new Round(DateTime.UtcNow, TimeSpan.FromMinutes(20));
    public async Task<Result<Round>> GetCurrentRound()
    {
        // Simulate async
        return await Task.Run(() =>
        {
            return _currentRound;
        });
    }
}

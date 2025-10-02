using API.DataAccess.Repositories;
using API.Models;
using Microsoft.AspNetCore.Mvc;

namespace API.Rounds;

[Route("[controller]")]
[ApiController]
public class RoundsController : ControllerBase
{
    private readonly RoundRepository _roundRepository;
    public RoundsController(RoundRepository roundRepository)
    {
        _roundRepository = roundRepository;
    }

    [HttpGet("end-time")]
    public async Task<IActionResult> GetEndTime()
    {
        var roundResult = await _roundRepository.GetCurrentRound();
        return roundResult.Match(
                round => Ok(round.EndTime), 
                error => error.AsActionResult()
        );        
    }
}

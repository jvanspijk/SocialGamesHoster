using API.DataAccess.Repositories;
using API.Models;
using Microsoft.AspNetCore.Mvc;

namespace API.Controllers;

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
        Result<Round> result = await _roundRepository.GetCurrentRound(); 

        if (!result.Ok)
        {
            return result.AsActionResult();
        }
        DateTime endTime = result.Value.EndTime;
        return Ok(endTime);
    }
}

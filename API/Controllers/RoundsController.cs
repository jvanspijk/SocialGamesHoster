using API.DataAccess.Repositories;
using API.Models;
using Microsoft.AspNetCore.Mvc;
using LanguageExt.Common;

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
        if (result.IsFaulted)
        {
            return result.ToActionResult();
        }
        var endTime = result.GetValueOrThrow().EndTime;
        return Ok(endTime);
    }
}

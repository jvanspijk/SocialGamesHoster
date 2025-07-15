using API.DataAccess.Repositories;
using API.Models;
using Microsoft.AspNetCore.Mvc;
using LanguageExt.Common;

namespace API.Controllers;

[Route("[controller]")]
[ApiController]
public class RoundController : ControllerBase
{
    private readonly RoundRepository _roundRepository;
    public RoundController(RoundRepository roundRepository)
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
        var endTime = result.ToObjectUnsafe().EndTime;
        return Ok(endTime);
    }
}

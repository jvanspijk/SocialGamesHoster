using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;

namespace API.Features.Abilities;

[Route("[controller]")]
[ApiController]
public class AbilitiesController : ControllerBase
{
    private readonly AbilityService _abilityService;
    public AbilitiesController(AbilityService abilityService)
    {
        _abilityService = abilityService;
    }

    [HttpGet]
    [Authorize]
    public async Task<IActionResult> GetAllAbilities()
    {
        var abilitiesResult = await _abilityService.GetAllAsync();
        return abilitiesResult.AsActionResult();
    }

    [HttpGet("{id}")]
    public async Task<IActionResult> GetAbility(int id)
    {
        var abilityResult = await _abilityService.GetAsync(id);        
        return abilityResult.AsActionResult();
    }
}

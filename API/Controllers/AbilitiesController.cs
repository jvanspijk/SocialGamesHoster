using API.DataAccess.Repositories;
using API.DTO;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

namespace API.Controllers;
[Route("[controller]")]
[ApiController]
public class AbilitiesController : ControllerBase
{
    private readonly AbilityRepository _abilityRepository;
    public AbilitiesController(AbilityRepository abilityRepository)
    {
        _abilityRepository = abilityRepository;
    }

    [HttpGet]
    [Authorize]
    public async Task<IActionResult> GetAllAbilities()
    {
        var abilitiesResult = await _abilityRepository.GetAllAbilitiesAsync();
        Console.WriteLine($"Abilities result: {abilitiesResult}");
        return abilitiesResult.AsActionResult();
    }

    [HttpGet("{id}")]
    public async Task<IActionResult> GetAbility(int id)
    {
        var abilityResult = await _abilityRepository.GetAbilityAsync(id);        
        return abilityResult.AsActionResult();       
    }
}

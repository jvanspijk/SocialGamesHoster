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
        return abilitiesResult.ToActionResult();
    }

    [HttpGet("{id}")]
    public async Task<IActionResult> GetAbility(int id)
    {
        var abilityResult = await _abilityRepository.GetAbilityAsync(id);
        if (abilityResult.IsFaulted)
        {
            return abilityResult.ToActionResult();
        }
        var ability = abilityResult.ToObjectUnsafe();
        if (ability == null)
        {
            return NotFound($"Ability with id {id} not found.");
        }
        return Ok(ability);
    }
}

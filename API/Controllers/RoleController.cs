using API.DataAccess.Repositories;
using API.Models;
using API.Validation;
using LanguageExt;
using LanguageExt.Common;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;

namespace API.Controllers;

[Route("[controller]")]
[ApiController]
public class RoleController : ControllerBase
{
    private readonly RoleRepository _roleRepository;
    private readonly PlayerRepository _playerRepository;
    public RoleController(RoleRepository roleRepository, PlayerRepository playerRepository)
    {
        _roleRepository = roleRepository;
        _playerRepository = playerRepository;
    }

    [HttpGet("{username}")]
    [Authorize]
    public async Task<IActionResult> GetRoleFromName(string username)
    {
        Claim? usernameClaim = User.FindFirst(JwtRegisteredClaimNames.Sub);
        if (usernameClaim == null)
        {
            return Unauthorized("User claim not found in token.");
        }
        if(usernameClaim.Value != username)
        {
            return Forbid("You are not allowed to access this user's role.");
        }

        Result<Role> roleResult = await _roleRepository.GetFromPlayerAsync(username);

        //Role test = roleResult.ToObjectUnsafe();
        //Console.WriteLine($"Role: {test.Name}, Id: {test.Id}");
        //Console.WriteLine($"Abilities: {string.Join(", ", test.AbilityAssociations.Select(a => a.Ability.Name))}");

        return roleResult.ToActionResult();      
    }
}

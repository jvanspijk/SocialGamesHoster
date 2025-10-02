using API.DTO;
using API.Models;
using API.Services;
using API.Validation;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;

namespace API.Controllers;

[ApiController]
[Route("[controller]")]
public class PlayersController : ControllerBase
{
    private readonly PlayerService _playerService;
    private readonly AuthService _authService;
    public PlayersController(PlayerService playerService, AuthService authService) 
    {
        _playerService = playerService;
        _authService = authService;
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] string username)
    {
        Result<Player> result = await _playerService.GetByNameAsync(username);
        if(!result.IsSuccess)
        {
            return result.AsActionResult();
        }
        string token = _authService.GeneratePlayerToken(username);
        return Ok(new { token });        
    }

    // GET /Players
    [HttpGet]
    public async Task<IActionResult> GetAll()
    {
        Result<List<Player>> result = await _playerService.GetAllAsync();
        return result.AsActionResult();
    }

    // GET /Players/{name}
    [HttpGet("{name}")]
    public async Task<IActionResult> GetByName(string name)
    {
        Result<Player> result = await _playerService.GetByNameAsync(name);
        return result.AsActionResult();
    }

    // POST /Players
    [HttpPost]
    public async Task<IActionResult> Create(string username)
    {
        Result<Player> user = await _playerService.CreateAsync(username);
        return user.AsActionResult();

    }

    // DELETE /Players/{id}
    [HttpDelete("{id}")]
    public async Task<IActionResult> Delete(int id)
    {
        Result<Player> result = await _playerService.DeleteAsync(id);
        return result.AsActionResult();
    }

    // PATCH /Players/{name}/Role
    [HttpPatch("{name}/Role")]
    [Authorize]
    public async Task<IActionResult> UpdateRole(string name, [FromBody] int roleId)
    {
        //bool isAdmin = _authService.IsAdmin(User);
        Result<Player> updatedPlayerResult = await _playerService.UpdateRoleAsync(name, roleId);
        return updatedPlayerResult.AsActionResult();
    }

    // GET /Players/{name}/Role
    [HttpGet("{name}/Role")]
    [Authorize]
    public async Task<IActionResult> GetRoleFromName(string name)
    {
        Claim? usernameClaim = User.FindFirst(JwtRegisteredClaimNames.Sub);
        Claim? roleClaim = User.FindFirst(ClaimTypes.Role);

        var canSeeResult = await _authService.CanSeePlayerAsync(usernameClaim, roleClaim, name);
        if (!canSeeResult.TryGetValue(out bool canSee))
        {
            return canSeeResult.AsActionResult();
        }

        if (!canSee)
        {
            return Unauthorized("You are not allowed to see this player.");
        }

        var roleResult = await _playerService.GetRoleFromPlayerAsync(name);
        if (!roleResult.TryGetValue(out var role))
        {
            return roleResult.AsActionResult();
        }            

        return Ok(new RoleDTO(role));
    }

    // GET /Players/{name}/VisiblePlayers
    [HttpGet("{name}/VisiblePlayers")]
    [Authorize]
    public async Task<IActionResult> GetPlayersVisibleToPlayer(string name)
    {
        Result<List<Player>> result = await _playerService.GetPlayersVisibleToPlayerAsync(name);
        return result.AsActionResult();
    }
}

using API.DataAccess.Repositories;
using API.DTO;
using API.Models;
using API.Services;
using LanguageExt.Common;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
using System;
using System.Collections.Generic;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;

namespace API.Controllers;

[ApiController]
[Route("[controller]")]
public class PlayersController : ControllerBase
{
    private readonly PlayerRepository _playerRepository;
    private readonly RoleRepository _roleRepository;
    private readonly AuthService _authService;
    public PlayersController(PlayerRepository playerRepository, 
        RoleRepository roleRepository, AuthService authService) 
    {
        _playerRepository = playerRepository;
        _roleRepository = roleRepository;
        _authService = authService;
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] string username)
    {
        Result<Player> result = await _playerRepository.GetPlayerByNameAsync(username);

        if (!result.IsSuccess)
        {
            Console.WriteLine($"Login failed for user {username}: {result.ToString()}");
            return result.ToActionResult();
        }

        string token = _authService.GeneratePlayerToken(username);
        return Ok(new { token });
    }

    // GET /Players
    [HttpGet]
    public async Task<IActionResult> GetAll()
    {
        Result<IEnumerable<Player>> result = await _playerRepository.GetPlayersAsync();
        return result.ToActionResult();
    }

    // GET /Players/{name}
    [HttpGet("{name}")]
    public async Task<IActionResult> GetByName(string name)
    {
        Result<Player> result = await _playerRepository.GetPlayerByNameAsync(name);
        return result.ToActionResult();
    }

    // POST /Players
    [HttpPost]
    public async Task<IActionResult> Create(string username)
    {
        Result<Player> user = await _playerRepository.AddPlayerAsync(username);
        return user.ToActionResult();

    }

    // DELETE /Players/{id}
    [HttpDelete("{id}")]
    public async Task<IActionResult> Delete(int id)
    {
        Result<Player> result = await _playerRepository.DeletePlayerAsync(id);
        return result.ToActionResult();
    }

    // PATCH /Players/{name}/Role
    [HttpPatch("{name}/Role")]
    [Authorize]
    public async Task<IActionResult> UpdateRole(string name, [FromBody] int roleId)
    {
        //bool isAdmin = _authService.IsAdmin(User);
        Result<Player> updatedPlayerResult = await _playerRepository.UpdateRole(name, roleId);
        return updatedPlayerResult.ToActionResult();
    }

    // GET /Players/{name}/Role
    [HttpGet("{name}/Role")]
    [Authorize]
    public async Task<IActionResult> GetRoleFromName(string name)
    {
        Claim? usernameClaim = User.FindFirst(JwtRegisteredClaimNames.Sub);
        Claim? roleClaim = User.FindFirst(ClaimTypes.Role);

        Result<bool> authResult = await _authService.CanSeePlayer(usernameClaim, roleClaim, name);
        if (!authResult.IsSuccess)
        {
            return authResult.ToActionResult();
        }

        bool canSeeRole = authResult.GetValueOrThrow();
        if(!canSeeRole)
        {
            return Unauthorized("You are not allowed to see this player.");
        }
       
        Result<Role> roleResult = await _roleRepository.GetRoleByPlayerNameAsync(name);
        if (!roleResult.IsSuccess)
        {
            return roleResult.ToActionResult();
        }

        return Ok(RoleDTO.FromModel(roleResult.GetValueOrThrow()));
    }

    // GET /Players/{name}/VisiblePlayers
    [HttpGet("{name}/VisiblePlayers")]
    [Authorize]
    public async Task<IActionResult> GetPlayersVisibleToPlayer(string name)
    {
        Result<List<Player>> result = await _playerRepository.GetPlayersVisibleToPlayerAsync(name);
        return result.ToActionResult();
    }
}

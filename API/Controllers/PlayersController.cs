using API.DataAccess.Repositories;
using API.Models;
using LanguageExt.Common;
using Microsoft.AspNetCore.Mvc;
using System;
using System.Collections.Generic;
using API.Services;

namespace API.Controllers;

[ApiController]
[Route("[controller]")]
public class PlayersController : ControllerBase
{
    private readonly PlayerRepository _playerRepository;
    private readonly AuthService _authService;
    public PlayersController(PlayerRepository playerRepository, AuthService authService) 
    {
        _playerRepository = playerRepository;
        _authService = authService;
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] string username)
    {
        Result<Player> result = await _playerRepository.GetPlayerAsync(username);

        if (!result.IsSuccess)
        {
            Console.WriteLine($"{result}");
            return result.ToActionResult();
        }

        string token = _authService.GeneratePlayerToken(username);
        return Ok(new { token });
    }

    // GET /User
    [HttpGet]
    public async Task<IActionResult> Get()
    {
        Result<IEnumerable<Player>> result = await _playerRepository.GetPlayersAsync();
        return result.ToActionResult();
    }

    // GET /User/{name}
    [HttpGet("{name}")]
    public async Task<IActionResult> GetUserByName(string name)
    {
        Result<Player> result = await _playerRepository.GetPlayerAsync(name);
        return result.ToActionResult();
    }

    // POST /User
    [HttpPost]
    public async Task<IActionResult> CreateUser(string username)
    {
        Result<Player> user = await _playerRepository.AddPlayerAsync(username);
        return user.ToActionResult();

    }
}

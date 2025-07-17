using API.DataAccess.Repositories;
using API.Models;
using LanguageExt.Common;
using Microsoft.AspNetCore.Mvc;
using System;
using System.Collections.Generic;

namespace API.Controllers;

[ApiController]
[Route("[controller]")]
public class PlayerController : ControllerBase
{
    private readonly PlayerRepository _playerRepository;
    public PlayerController(PlayerRepository playerRepository) 
    {
        _playerRepository = playerRepository;        
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

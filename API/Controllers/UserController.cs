using API.DataAccess.Repositories;
using API.Models;
using LanguageExt.Common;
using Microsoft.AspNetCore.Mvc;
using System;
using System.Collections.Generic;

namespace API.Controllers;

[ApiController]
[Route("[controller]")]
public class UserController : ControllerBase
{
    private readonly UserRepository _userRepository;
    public UserController(UserRepository userRepository) 
    {
        _userRepository = userRepository;        
    }

    // GET /User
    [HttpGet]
    public async Task<IActionResult> Get()
    {
        Result<IEnumerable<User>> result = await _userRepository.GetUsersAsync();
        return result.ToActionResult();
    }

    // GET /User/{name}
    [HttpGet("{name}")]
    public async Task<IActionResult> GetUserByName(string name)
    {
        Result<User> result = await _userRepository.GetUserAsync(name);
        return result.ToActionResult();
    }

    // POST /User
    [HttpPost]
    public async Task<IActionResult> CreateUser(string username)
    {
        Result<User> user = await _userRepository.AddUserAsync(username);
        return user.ToActionResult();

    }
}

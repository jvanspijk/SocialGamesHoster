using API.Models;
using API.Services;
using Microsoft.AspNetCore.Mvc;
using System;

namespace API.Controllers;

[ApiController]
[Route("[controller]")]
public class UserController : ControllerBase
{
    private readonly UserService _userService;
    public UserController(UserService userService) 
    {
        _userService = userService;        
    }

    // Post /User/login/{name}
    [HttpPost("login/{name}", Name = "LoginUser")]
    public ActionResult<User> LogIn(string name)
    {
        try
        {
            User user = _userService.LogIn(name);
            return Ok(user);
        }
        catch (Exception ex)
        {
            return BadRequest(new { error = ex.Message });
        }                
    }

    // GET /User
    [HttpGet(Name = "GetAllUsers")]
    public IEnumerable<User> Get()
    {
        return _userService.GetUsers();
    }

    // GET /User/{name}
    [HttpGet("{name}", Name = "GetUserByName")]
    public ActionResult<User> GetUserByName(string name)
    {
        var user = _userService.GetUser(name);
        if (user == null)
        {
            return NotFound($"User {name} not found.");
        }

        return Ok(user);
    }

    // POST /User
    [HttpPost]
    public ActionResult<User> CreateUser(string username)
    {
        try
        {
            var user = _userService.AddUser(username);
            return Ok(user);
        }
        catch (Exception ex)
        {
            return BadRequest(ex);
        }

    }
}

using Microsoft.AspNetCore.Mvc;
using SocialGamesHoster.Models;

namespace SocialGamesHoster.Controllers;

[ApiController]
[Route("[controller]")]
public class UserController : ControllerBase
{
    private readonly List<User> _users = [];
    public UserController() 
    {
        List<string> testUserNames = ["Bob", "Jan", "Kees"];
        foreach(string name in testUserNames)
        {
            _users.Add(new User { Name = name });
        }
        
    }

    // GET /User
    [HttpGet(Name = "GetAllUsers")]
    public IEnumerable<User> Get()
    {
        return _users;
    }

    // GET /User/{name}
    [HttpGet("{name}", Name = "GetUserByName")]
    public ActionResult<User> GetUserByName(string name)
    {
        var user = _users.FirstOrDefault(u => u.Name.ToLower() == name.ToLower());
        if (user == null)
        {
            return NotFound($"User {name} not found.");
        }

        return Ok(user);
    }

    // POST /User
    [HttpPost]
    public ActionResult<User> CreateUser([FromBody] User newUser)
    {
        if (string.IsNullOrWhiteSpace(newUser.Name))
        {
            return BadRequest("User name cannot be empty.");
        }

        if (_users.Any(u => u.Name.ToLower() == newUser.Name.ToLower()))
        {
            return BadRequest($"User with name '{newUser.Name}' already exists.");
        }

        _users.Add(newUser);
        return Ok(newUser);
    }
}

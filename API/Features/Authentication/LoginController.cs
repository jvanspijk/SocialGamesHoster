using API.Features.Authentication.Requests;
using API.Features.Players;
using API.Models;
using Microsoft.AspNetCore.Mvc;

namespace API.Features.Authentication;
[Route("[controller]")]
[ApiController]
public class LoginController : ControllerBase
{
    private readonly AuthService _authService;
    private readonly PlayerService _playerService;
    public LoginController(AuthService authService, PlayerService playerService)
    {
        _authService = authService;   
        _playerService = playerService;
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] string username)
    {
        Result<Player> result = await _playerService.GetByNameAsync(username);
        if (!result.IsSuccess)
        {
            return result.AsActionResult();
        }
        string token = _authService.GeneratePlayerToken(username);
        return Ok(new { token });
    }

    [HttpPost("admin/login")]
    public async Task<IActionResult> AdminLogin([FromBody] AdminLoginRequest request)
    {
        if (!ModelState.IsValid)
        {
            return BadRequest(ModelState);
        }

        if (!_authService.AdminCredentialsAreValid(request.Username, request.Password))
        {
            return Unauthorized("Invalid credentials.");
        }

        string token = await Task.FromResult(_authService.GenerateAdminToken(request.Username));
        return Ok(new { token });
    }
}

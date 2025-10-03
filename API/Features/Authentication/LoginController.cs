using API.Domain;
using API.Domain.Models;
using API.Features.Authentication.Requests;
using API.Features.Authentication.Responses;
using API.Features.Players;
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

    [HttpPost]
    public async Task<IActionResult> Login([FromBody] PlayerLoginRequest request)
    {
        string name = request.Name;
        Result<Player> result = await _playerService.GetByNameAsync(name);
        if (!result.IsSuccess)
        {
            return result.AsActionResult();
        }
        string token = _authService.GeneratePlayerToken(name);
        return Ok(new LoginTokenResponse(token));
    }

    [HttpPost("admin/login")]
    public async Task<IActionResult> AdminLogin([FromBody] AdminLoginRequest request)
    {
        if (!ModelState.IsValid)
        {
            return BadRequest(ModelState);
        }

        bool credentialsAreValid = await _authService.AdminCredentialsAreValid(request.Username, request.Password);
        if (!credentialsAreValid)
        {
            return Unauthorized("Invalid credentials.");
        }

        string token = _authService.GenerateAdminToken(request.Username);
        return Ok(new LoginTokenResponse(token));
    }
}

using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.GameSessions.Authentication.Responses;
using Microsoft.AspNetCore.Mvc;
using API.Features.Authentication.Requests;

namespace API.Features.Authentication;

[Route("[controller]")]
[ApiController]
public class LoginController : ControllerBase
{
    private readonly AuthService _authService;
    private readonly PlayerRepository _playerRepository;
    public LoginController(AuthService authService, PlayerRepository playerService)
    {
        _authService = authService;   
        _playerRepository = playerService;
    }

    [HttpPost]
    public async Task<IActionResult> Login([FromBody] PlayerLoginRequest request)
    {
        string name = request.Name;
        Player? result = await _playerRepository.GetByNameAsync(name, request.GameId);
       
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

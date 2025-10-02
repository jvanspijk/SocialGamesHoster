using Microsoft.AspNetCore.Mvc;
using API.Services;
using API.Models.Requests;

namespace API.Controllers;
[Route("[controller]")]
[ApiController]
public class AdminController : ControllerBase
{
    private readonly AuthService _authService;
    public AdminController(AuthService authService)
    {
        _authService = authService;        
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] AdminLoginRequestDTO request)
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

using API.DTO;
using Microsoft.AspNetCore.Mvc;

namespace API.Controllers;
[Route("[controller]")]
[ApiController]
public class AdminController : ControllerBase
{
    private const string ADMIN_USERNAME = "admin"; // don't worry, this is just for testing
    private const string ADMIN_PASSWORD = "admin"; // don't worry, this is just for testing

    private readonly JwtTokenService _tokenService;

    public AdminController(JwtTokenService tokenService)
    {
        _tokenService = tokenService;
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] AdminLoginRequestDTO request)
    {
        if (request.Username != ADMIN_USERNAME || request.Password != ADMIN_PASSWORD)
        {
            return Unauthorized("Invalid credentials.");
        }

        string token = _tokenService.GenerateAdminToken(request.Username);
        return Ok(new { token });
    }
}

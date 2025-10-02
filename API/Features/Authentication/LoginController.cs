using Microsoft.AspNetCore.Mvc;
using API.Features.Authentication.Requests;

namespace API.Features.Authentication;
[Route("[controller]")]
[ApiController]
public class LoginController : ControllerBase
{
    private readonly AuthService _authService;
    public LoginController(AuthService authService)
    {
        _authService = authService;        
    }

    [HttpPost("login")]
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

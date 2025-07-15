using API.DataAccess.Repositories;
using API.Models;
using LanguageExt.Common;
using Microsoft.AspNetCore.Mvc;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace API.Controllers;
[Route("[controller]")]
[ApiController]
public class AuthController : ControllerBase
{
    private readonly UserRepository _userRepository;
    public AuthController(UserRepository userRepository)
    {
        _userRepository = userRepository;
    }

    [HttpPost("login")]
    public async Task<IActionResult> Login([FromBody] string username)
    {
        Result<User> user = await _userRepository.GetUserAsync(username);
        if (!user.IsSuccess)
        {
            return NotFound("User not found.");
        }
        string token = GenerateJwtToken(username);
        return Ok(new { token });
    }

    private string GenerateJwtToken(string username)
    {
        Claim[] claims =
        {
            new(JwtRegisteredClaimNames.Sub, username),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString())
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes("Social-games-hoster_JWT_Security_Key")); // TODO: env variable
        var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: "http://localhost:8080", // TODO: env variable
            audience: "http://localhost:8081", //TODO: env variable
            claims: claims,
            expires: DateTime.Now.AddMinutes(10),
            signingCredentials: creds);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }
}

using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using Microsoft.IdentityModel.Tokens;

public class JwtTokenService
{
    // TODO: these should be loaded from environment variables
    private const string JwtSecurityKey = "Social-games-hoster_JWT_Security_Key";
    private const string Issuer = "http://localhost:8080";
    private const string Audience = "http://localhost:8081";

    /// <summary>
    /// Generates a JWT token for a user with the 'player' role.
    /// </summary>
    public string GeneratePlayerToken(string username)
    {
        return GenerateToken(username, "player");
    }

    /// <summary>
    /// Generates a JWT token for a user with the 'admin' role.
    /// </summary>
    public string GenerateAdminToken(string username)
    {
        return GenerateToken(username, "admin");
    }

    /// <summary>
    /// Generates a JWT token
    /// </summary>
    private string GenerateToken(string username, string role)
    {
        Claim[] claims =
        {
            new(JwtRegisteredClaimNames.Sub, username),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString()),
            new(ClaimTypes.Role, role)
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(JwtSecurityKey));
        var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: Issuer,
            audience: Audience,
            claims: claims,
            expires: DateTime.Now.AddMinutes(10),
            signingCredentials: creds);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }
}
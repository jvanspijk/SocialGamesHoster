using API;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace API.Features.Auth;

public class AuthService(PlayerRepository playerRepository)
{
    // TODO: these should be loaded from environment variables
    private const string _jwtSecurityKey = "Social-games-hoster_JWT_Security_Key";
    private const string _jwtIssuer = "http://localhost:9090";
    private const string _jwtAudience = "http://localhost:9091";
    private const string _playerTokenName = "player_token";
    private const string _adminTokenName = "admin_token";

    private readonly string _adminUserName = "admin";
    private readonly string _adminPassword = "admin";
    private const string _adminRoleName = "admin";

    private readonly PlayerRepository _playerRepository = playerRepository;

    public Task<bool> AdminCredentialsAreValid(string username, string passwordHash)
    {
        // Pretend this is an async operation because it probably will be in the future
        return Task.FromResult(username == _adminUserName && passwordHash == _adminPassword);
    }

    /// <summary>
    /// Generates a JWT token for a user with the 'player' role.
    /// </summary>
    public string GeneratePlayerToken(int userId, string name, int? roleId)
    {
        var claims = new List<Claim>
        {
            new(JwtRegisteredClaimNames.Sub, userId.ToString()),
            new(JwtRegisteredClaimNames.Name, name),
            new("role", roleId.ToString() ?? ""),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString()),
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_jwtSecurityKey));
        var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: _jwtIssuer,
            audience: _jwtAudience,
            claims: claims,
            expires: DateTime.Now.AddHours(12),
            signingCredentials: creds);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    /// <summary>
    /// Generates a JWT token for a user with the 'admin' role.
    /// </summary>
    public string GenerateAdminToken()
    {
        var claims = new List<Claim>
        {
            new(JwtRegisteredClaimNames.Sub, "0"),
            new(JwtRegisteredClaimNames.UniqueName, _adminUserName),
            new("role", _adminRoleName),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString())
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_jwtSecurityKey));
        var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: _jwtIssuer,
            audience: _jwtAudience,
            claims: claims,
            expires: DateTime.Now.AddMonths(1),
            signingCredentials: creds);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    public async Task<Result<bool>> CanSeePlayerAsync(HttpRequest request, int targetPlayerId)
    {
        var adminCheck = IsAdmin(request);
        if (adminCheck.IsSuccess && adminCheck.Value)
        {
            return true;
        }

        var playerClaimsResult = GetPlayerClaims(request);
        if (playerClaimsResult.IsFailure)
        {
            return playerClaimsResult.Error;
        }

        (int userId, int? _) = playerClaimsResult.Value;

        Player? sourcePlayer = await _playerRepository.GetAsync(userId);
        if (sourcePlayer == null)
        {
            return Errors.ResourceNotFound(nameof(Player), nameof(Player.Name), userId.ToString());
        }

        Player? targetPlayer = await _playerRepository.GetAsync(targetPlayerId);
        if (targetPlayer == null)
        {
            return Errors.ResourceNotFound(nameof(Player), nameof(Player.Id), targetPlayerId.ToString());
        }

        return await _playerRepository.IsVisibleToPlayerAsync(sourcePlayer, targetPlayer);
    }

    public Result<bool> IsAdmin(HttpRequest request)
    {
        string? token = request.Cookies[_adminTokenName];
        if (string.IsNullOrEmpty(token))
        {
            return Errors.MissingClaims("Admin token is missing.");
        }

        try
        {
            var handler = new JwtSecurityTokenHandler();
            var jwtToken = handler.ReadJwtToken(token);

            var adminClaim = jwtToken.Claims.FirstOrDefault(c =>
                c.Type == JwtRegisteredClaimNames.Name || c.Type == "unique_name");

            if (adminClaim?.Value == _adminUserName)
            {
                return true;
            }

            return false;
        }
        catch 
        {
            return Errors.InvalidToken("Admin token is malformed.");
        }
    }

    public static Result<(int, int?)> GetPlayerClaims(HttpRequest request)
    {
        Console.WriteLine($"Cookies found: {string.Join("; ", request.Cookies.Select(c => $"{c.Key}={c.Value}"))}");
        if (!request.Cookies.TryGetValue(_playerTokenName, out string? token) || string.IsNullOrEmpty(token))
        {
            return Errors.MissingClaims("Player token cookie is missing.");
        }

        try
        {
            var handler = new JwtSecurityTokenHandler();
            var jwtToken = handler.ReadJwtToken(token);

            var subClaim = jwtToken.Claims.FirstOrDefault(c => c.Type == JwtRegisteredClaimNames.Sub || c.Type == "nameid");
            if (subClaim == null || !int.TryParse(subClaim.Value, out int userId))
            {
                return Errors.InvalidToken("User ID claim is missing or invalid.");
            }

            var roleClaim = jwtToken.Claims.FirstOrDefault(c => c.Type == "role" || c.Type == ClaimTypes.Role);
            int? roleId = int.TryParse(roleClaim?.Value, out int rId) ? rId : null;

            return (userId, roleId);
        }
        catch
        {
            return Errors.InvalidToken("Player token is malformed.");
        }
    }
}
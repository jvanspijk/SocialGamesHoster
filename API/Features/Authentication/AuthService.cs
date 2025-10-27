using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace API.Features.Authentication;

public class AuthService
{
    // TODO: these should be loaded from environment variables
    private const string _jwtSecurityKey = "Social-games-hoster_JWT_Security_Key";
    private const string _jwtIssuer = "http://localhost:9090";
    private const string _jwtAudience = "http://localhost:9091";

    private readonly string _adminUserName;
    private readonly string _adminPassword;
    private const string _adminRoleName = "admin";

    private readonly PlayerRepository _playerRepository;

    public AuthService(PlayerRepository playerRepository)
    {
        // TODO: replace with environment variables
        //if (!IniParser.ParseIniFile("settings.ini").TryGetValue("Admin", out var adminSettings))
        //{
        //    throw new InvalidOperationException("Can't find admin header in settings.ini.");
        //}

        //if (!adminSettings.TryGetValue("admin_username", out string? username) || !adminSettings.TryGetValue("admin_password", out string? password))
        //{
        //    throw new InvalidOperationException("Admin username or password not set in settings.ini.");
        //}

        _adminUserName = "admin";
        _adminPassword = "admin";

        _playerRepository = playerRepository;
    }

    public Task<bool> AdminCredentialsAreValid(string username, string passwordHash)
    {
        // Pretend this is an async operation because it probably will be in the future
        return Task.FromResult(username == _adminUserName && passwordHash == _adminPassword);
    }

    /// <summary>
    /// Generates a JWT token for a user with the 'player' role.
    /// </summary>
    public string GeneratePlayerToken(int userId, int? roleId)
    {
        var claims = new List<Claim>
        {
            new(JwtRegisteredClaimNames.Sub, userId.ToString()),
            new(JwtRegisteredClaimNames.Jti, Guid.NewGuid().ToString()),
            new(ClaimTypes.Role, roleId.ToString() ?? "")
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_jwtSecurityKey));
        var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: _jwtIssuer,
            audience: _jwtAudience,
            claims: claims,
            expires: DateTime.Now.AddHours(24),
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
            new(ClaimTypes.Role, _adminRoleName)
        };

        var key = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(_jwtSecurityKey));
        var creds = new SigningCredentials(key, SecurityAlgorithms.HmacSha256);

        var token = new JwtSecurityToken(
            issuer: _jwtIssuer,
            audience: _jwtAudience,
            claims: claims,
            expires: DateTime.Now.AddHours(24),
            signingCredentials: creds);

        return new JwtSecurityTokenHandler().WriteToken(token);
    }  

    public async Task<Result<bool>> CanSeePlayerAsync(ClaimsPrincipal userClaim, int targetPlayerId)
    {
        var adminCheck = IsAdmin(userClaim);
        if (adminCheck.IsFailure)
        {
            return adminCheck.Error;
        }

        bool isAdmin = adminCheck.Value;
        if (isAdmin)
        {
            return true;
        }

        var playerClaimsResult = GetPlayerClaims(userClaim);
        if (playerClaimsResult.IsFailure)
        {
            return playerClaimsResult.Error;
        }
        (int userId, int roleId) = playerClaimsResult.Value;

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

    public Result<bool> IsAdmin(ClaimsPrincipal user)
    {
        var roleClaim = user.Claims.FirstOrDefault(c => c.Type == ClaimTypes.Role);
        if (roleClaim == null)
        {
            return Errors.MissingClaims("Role claim is missing.");
        }
        return string.Equals(roleClaim.Value, _adminRoleName, StringComparison.Ordinal);
    }

    private static Result<(int, int)> GetPlayerClaims(ClaimsPrincipal playerClaim)
    {
        Claim? playerIdClaim = playerClaim.FindFirst(JwtRegisteredClaimNames.Sub);
        Claim? roleClaim = playerClaim.FindFirst(ClaimTypes.Role);

        if (roleClaim == null || playerIdClaim == null)
        {
            return Errors.MissingClaims("Role or username claim is missing.");
        }

        if (!int.TryParse(playerIdClaim.Value, out int userId))
        {
            return Errors.InvalidToken($"User ID claim `{playerIdClaim?.Value}` is not a valid integer.");
        }

        if(!int.TryParse(roleClaim.Value, out int roleId))
        {
            return Errors.InvalidToken($"Role ID claim `{roleClaim?.Value}` is not a valid integer.");
        }

        return (userId, roleId);
    }
}
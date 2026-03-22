using API.DataAccess;
using API.Domain;
using API.Domain.Entities;
using API.Domain.Validation;
using Microsoft.IdentityModel.Tokens;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;

namespace API.Features.Auth;

public class AuthService(IRepository<Player> playerRepository)
{
    // TODO: these should be loaded from environment variables
    private const string _jwtSecurityKey = "Social-games-hoster_JWT_Security_Key";
    private const string _jwtIssuer = "http://localhost:9090";
    private const string _jwtAudience = "http://localhost:9091";
    private const string _playerTokenName = "player_token";
    private const string _adminTokenName = "admin_token";

    private static readonly string _adminUserName = "admin";
    private static readonly string _adminPassword = "admin";
    private static readonly string _adminRoleName = "admin";

    private readonly IRepository<Player> _playerRepository = playerRepository;

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

        Player? sourcePlayer = await _playerRepository.GetWithTrackingAsync(userId);
        if (sourcePlayer == null)
        {
            return Errors.ResourceNotFound(nameof(Player), nameof(Player.Name), userId.ToString());
        }

        Player? targetPlayer = await _playerRepository.GetWithTrackingAsync(targetPlayerId);
        if (targetPlayer == null)
        {
            return Errors.ResourceNotFound(nameof(Player), nameof(Player.Id), targetPlayerId.ToString());
        }

        // TODO: implement as an extension method on PlayerRepository that checks the CanSee and CanBeSeenBy relationships instead of loading the full player entities and their relationships into memory
        //return await _playerRepository.Query()
        //   .Where(p => p.Id == sourcePlayer.Id)
        //   .SelectMany(p => p.CanSee)
        //   .AnyAsync(p => p.Id == targetPlayer.Id);

        // This will fail because the CanSee collection is not being eagerly loaded, but it illustrates the intended logic without needing to implement a custom repository method
        // return sourcePlayer.CanSee.Any(p => p.Id == targetPlayer.Id);

        // return await _playerRepository.IsVisibleToPlayerAsync(sourcePlayer, targetPlayer);
        return true;
    }

    public static Result<bool> IsAdmin(HttpRequest request)
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
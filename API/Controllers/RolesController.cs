using API.DataAccess.Repositories;
using API.DTO;
using API.Models;
using API.Services;
using API.Validation;
using LanguageExt;
using LanguageExt.Common;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;

namespace API.Controllers;

[Route("[controller]")]
[ApiController]
public class RolesController : ControllerBase
{

    private readonly RoleRepository _roleRepository;
    private readonly PlayerRepository _playerRepository;
    public RolesController(RoleRepository roleRepository, PlayerRepository playerRepository)
    {
        _roleRepository = roleRepository;
        _playerRepository = playerRepository;
    }

    [HttpGet]
    [Authorize]
    public async Task<IActionResult> GetAllRoles()
    {
        Claim? roleClaim = User.FindFirst(ClaimTypes.Role);

        if (roleClaim == null)
        {
            return BadRequest("Role claim are missing.");
        }


        bool isAdmin = string.Equals(roleClaim?.Value, "admin", StringComparison.OrdinalIgnoreCase);
        if(!isAdmin)
        {
            return Unauthorized("You are not allowed to see all roles.");
        }
        
        Result<ICollection<RoleDTO>> rolesResult;
        rolesResult = await _roleRepository.GetAllAsync();
        return rolesResult.ToActionResult();
    }


    [HttpGet("{username}")]
    [Authorize]
    public async Task<IActionResult> GetRoleFromName(string username)
    {
        Claim? usernameClaim = User.FindFirst(JwtRegisteredClaimNames.Sub);
        Claim? roleClaim = User.FindFirst(ClaimTypes.Role);

        if(usernameClaim == null || roleClaim == null)
        {
            return BadRequest("Claims are missing.");
        }

        Result<RoleDTO> targetRoleResult;

        bool isAdmin = roleClaim?.Value == "admin";

        //Admins can see any role
        if (isAdmin) 
        {
            targetRoleResult = await _roleRepository.GetFromPlayerAsync(username);
            return targetRoleResult.ToActionResult();
        }

        string requestingUsername = usernameClaim.Value;
        // Players can always see their own role
        if (requestingUsername.Equals(username, StringComparison.CurrentCultureIgnoreCase))
        {
            targetRoleResult = await _roleRepository.GetFromPlayerAsync(username);
            return targetRoleResult.ToActionResult();
        }

        Result<RoleDTO> ownRoleResult = await _roleRepository.GetFromPlayerAsync(requestingUsername);
        if (!ownRoleResult.IsSuccess)
        {
            return ownRoleResult.ToActionResult();
        }
        targetRoleResult = await _roleRepository.GetFromPlayerAsync(username);
        if (!targetRoleResult.IsSuccess)
        {
            return targetRoleResult.ToActionResult();
        }
        RoleDTO ownRole = ownRoleResult.ToObjectUnsafe();
        RoleDTO targetRole = targetRoleResult.ToObjectUnsafe();
        bool canSeeRole = ownRole.RolesVisibleToRole.Contains(targetRole.Id);

        if (canSeeRole)
        {
            return targetRoleResult.ToActionResult();
        }

        Result<Player> requestingPlayerResult = await _playerRepository.GetPlayerAsync(requestingUsername);
        Result<Player> targetPlayerResult = await _playerRepository.GetPlayerAsync(username);
        if (!requestingPlayerResult.IsSuccess || !targetPlayerResult.IsSuccess)
        {
            return BadRequest("Can't retrieve players.");
        }

        ICollection<int> playersVisibleToRequestingPlayer = requestingPlayerResult.ToObjectUnsafe().PlayersVisibleToPlayer;
        int targetPlayerId = targetPlayerResult.ToObjectUnsafe().Id;
        if(playersVisibleToRequestingPlayer.Contains(targetPlayerId))
        {
            return targetRoleResult.ToActionResult();
        }

        else
        {
            return BadRequest("You are not allowed to see this role.");
        }
    }
}

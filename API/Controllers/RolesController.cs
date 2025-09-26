using API.DataAccess.Repositories;
using API.DTO;
using API.Models;
using API.Services;
using API.Validation;
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
    private readonly AuthService _authService;
    public RolesController(RoleRepository roleRepository, 
        PlayerRepository playerRepository, AuthService authService)
    {
        _roleRepository = roleRepository;
        _playerRepository = playerRepository;
        _authService = authService;
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
        
        var rolesResult = await _roleRepository.GetAllAsync();
        return rolesResult.AsActionResult();
    }
}

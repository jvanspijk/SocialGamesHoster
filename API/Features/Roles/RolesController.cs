using API.Models;
using API.Validation;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;

namespace API.Features.Roles;

[Route("[controller]")]
[ApiController]
public class RolesController : ControllerBase
{
    private readonly RoleService _roleService;

    public RolesController(RoleService roleService)
    {
        _roleService = roleService;
    }

    [HttpGet]
    [Authorize]
    public async Task<IActionResult> GetAllRoles()
    {
        Claim? roleClaim = User.FindFirst(ClaimTypes.Role);

        if (roleClaim == null)
        {
            return BadRequest("Role claim is missing.");
        }

        bool isAdmin = string.Equals(roleClaim?.Value, "admin", StringComparison.OrdinalIgnoreCase);
        if(!isAdmin)
        {
            return Unauthorized("You are not allowed to see all roles.");
        }        
        
        var rolesResult = await _roleService.GetAllAsync();
        return rolesResult.AsActionResult();
    }
}

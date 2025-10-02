using API.Domain;
using API.Features.Roles.Requests;
using API.Features.Roles.Responses;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Mvc;
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

    // POST /Roles
    [HttpPost]
    public async Task<IActionResult> Create(CreateRoleRequest roleRequest)
    {
        var result = await _roleService.CreateAsync(roleRequest.Name, roleRequest.Description);
        return result.AsActionResult();

    }

    // DELETE /Players/{id}
    [HttpDelete("{id}")]
    public async Task<IActionResult> Delete(int id)
    {
        Result<bool> result = await _roleService.DeleteAsync(id);
        return result.AsActionResult();
    }
}

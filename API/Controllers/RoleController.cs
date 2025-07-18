using API.DataAccess.Repositories;
using API.DTO;
using API.Models;
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
public class RoleController : ControllerBase
{
    private readonly RoleRepository _roleRepository;
    public RoleController(RoleRepository roleRepository)
    {
        _roleRepository = roleRepository;
    }

    /// <summary>
    /// Currently, you can only retrieve your own role.
    /// </summary>
    /// <param name="username"></param>
    /// <returns></returns>
    [HttpGet("{username}")]
    [Authorize]
    public async Task<IActionResult> GetRoleFromName(string username)
    {
        Claim? usernameClaim = User.FindFirst(JwtRegisteredClaimNames.Sub);

        if (usernameClaim == null)
        {
            return Unauthorized("User claim not found in token.");
        }
        if(usernameClaim.Value != username)
        {
            return Forbid("You are not allowed to access this user's role.");
        }

        Result<RoleDTO> roleResult = await _roleRepository.GetFromPlayerAsync(username);        

        return roleResult.ToActionResult();      
    }
}

using API.DataAccess.Repositories;
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
    private readonly UserRepository _userRepository;
    public RoleController(RoleRepository roleRepository, UserRepository userRepository)
    {
        _roleRepository = roleRepository;
        _userRepository = userRepository;
    }

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

        Result<User> userResult = await _userRepository.GetUserAsync(username);

        if (!userResult.IsSuccess) return userResult.ToActionResult();

        User user = userResult.ToObjectUnsafe();

        if (user.Role == null) return NotFound("User does not have a role assigned.");

        Result<Role> roleResult = await _roleRepository.GetFromIdAsync(user.Role.Id);
        return roleResult.ToActionResult();
    }
}

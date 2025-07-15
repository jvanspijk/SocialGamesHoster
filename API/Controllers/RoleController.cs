using API.DataAccess.Repositories;
using API.Models;
using API.Validation;
using LanguageExt;
using LanguageExt.Common;
using Microsoft.AspNetCore.Authorization;
using Microsoft.AspNetCore.Http;
using Microsoft.AspNetCore.Mvc;

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
        Result<User> userResult = await _userRepository.GetUserAsync(username);

        if (!userResult.IsSuccess) return userResult.ToActionResult();

        User user = userResult.Match(
            Succ: u => u,
            Fail: _ => throw new ExceptionalException(
                "An error occurred while retrieving the user. " +
                "This should never happen.",
                new NotFoundException("User retrieval failed.")
            )
        );

        if (user.Role == null) return NotFound("User does not have a role assigned.");

        Result<Role> roleResult = await _roleRepository.GetFromIdAsync(user.Role.Id);
        return roleResult.ToActionResult();
    }
}

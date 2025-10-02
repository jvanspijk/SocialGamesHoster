using API.DataAccess.Repositories;
using API.Features.Roles.Responses;
using API.Models;
using API.Validation;

namespace API.Features.Roles;
public class RoleService
{
    private readonly RoleRepository _roleRepository;
    public RoleService(RoleRepository roleRepository)
    {
        _roleRepository = roleRepository;
    }
    public async Task<Result<List<RoleResponse>>> GetAllAsync()
    {
        List<Role> roles = await _roleRepository.GetAllAsync();
        return roles.Select(role => new RoleResponse(role)).ToList();
    }
    public async Task<Result<RoleResponse>> GetRoleAsync(int id)
    {
        Role? role = await _roleRepository.GetByIdAsync(id);
        if(role is null)
        {
            return Errors.ResourceNotFound("Role", id);
        }
        return new RoleResponse(role);
    }
}

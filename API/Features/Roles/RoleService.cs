using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using API.Features.Roles.Responses;

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

    public async Task<Result<RoleResponse>> CreateAsync(string name, string description)
    {
        Role newRole = new() { Name = name, Description = description };
        Role? createdRole = await _roleRepository.CreateAsync(newRole);
        if (createdRole is null)
        {
            return Errors.FailedToCreate("Role", name);
        }
        return new RoleResponse(createdRole);
    }

    public async Task<Result<bool>> DeleteAsync(int id)
    {
        Role? role = await _roleRepository.GetByIdAsync(id);
        if (role is null)
        {
            return Errors.ResourceNotFound("Role", id);
        }
        await _roleRepository.DeleteAsync(role);
        return true;
    }
}

using API.DataAccess.Repositories;
using API.Features.Roles.Responses;
using API.Models;

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
        var repositoryResult = await _roleRepository.GetAllAsync();
        return repositoryResult.Map(roles =>
            roles.Select(static r => new RoleResponse(r)).ToList());
        
    }
    public async Task<Result<RoleResponse>> GetRoleAsync(int id)
    {
        var repositoryResult = await _roleRepository.GetAsync(id);
        return repositoryResult.Map(static r => new RoleResponse(r));
    }
}

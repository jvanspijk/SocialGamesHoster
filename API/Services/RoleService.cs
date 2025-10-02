using API.DataAccess.Repositories;
using API.Models;
using API.Models.Requests;

namespace API.Services;
public class RoleService
{
    private readonly RoleRepository _roleRepository;
    public RoleService(RoleRepository roleRepository)
    {
        _roleRepository = roleRepository;
    }
    public async Task<Result<List<RoleDTO>>> GetAllAsync()
    {
        var repositoryResult = await _roleRepository.GetAllAsync();
        return repositoryResult.Map(roles =>
            roles.Select(static r => new RoleDTO(r)).ToList());
        
    }
    public async Task<Result<RoleDTO>> GetRoleAsync(int id)
    {
        var repositoryResult = await _roleRepository.GetAsync(id);
        return repositoryResult.Map(static r => new RoleDTO(r));
    }
}

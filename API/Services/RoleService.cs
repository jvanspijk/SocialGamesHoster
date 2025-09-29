using API.DataAccess.Repositories;
using API.DTO;

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
        var rolesResult = await _roleRepository.GetAllAsync();
        return rolesResult.Map(static roles =>
            roles.Select(static r => new RoleDTO(r)).ToList());
    }
    public async Task<Result<RoleDTO>> GetRoleAsync(int id)
    {
        var role = await _roleRepository.GetAsync(id);
        return role.Map(static r => new RoleDTO(r));
    }
}

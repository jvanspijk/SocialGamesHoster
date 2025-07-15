using API.Models;
using API.Validation;
using LanguageExt.Common;

namespace API.DataAccess.Repositories;

public class RoleRepository
{
    private List<Role> _roles;

    public RoleRepository() {
        _roles = new List<Role>(3);
        _roles = [
            new Role { Id = 1, Name = "Rol 1", Description = "Je kan eigenlijk niet zo veel." },
            new Role { Id = 2, Name = "Rol 2", Description = "Je kan nog minder." },
            new Role { Id = 3, Name = "Rol 3", Description = "Heel nuttig ben je niet."},
        ];
    }

    public async Task<Result<Role>> GetFromIdAsync(int id)
    {
        // Simulate async
        return await Task.Run(() =>
        {
            Role? role = _roles.FirstOrDefault(r => r.Id == id);

            if (role == null)
            {
                var exception = new NotFoundException($"Role with id {id} not found.");
                return new Result<Role>(exception);
            }
            return role;
        });
    }
}
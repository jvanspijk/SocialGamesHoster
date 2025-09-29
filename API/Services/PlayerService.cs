using API.DataAccess.Repositories;
using API.Models;

namespace API.Services;

public class PlayerService
{
    private readonly PlayerRepository _playerRepository;
    private readonly RoleRepository _roleRepository;
    public PlayerService(PlayerRepository playerRepository, RoleRepository roleRepository)
    {
        _playerRepository = playerRepository;
        _roleRepository = roleRepository;
    }

    public async Task<Result<Player>> GetAsync(int id)
    {
        return await _playerRepository.GetAsync(id);
    }

    public async Task<Result<Player>> GetByNameAsync(string username)
    {
        return await _playerRepository.GetByNameAsync(username);
    }

    public async Task<Result<Player>> CreateAsync(string username)
    {
        Player newPlayer = new Player
        {
            Name = username
        };
        return await _playerRepository.CreateAsync(newPlayer);
    }

    public async Task<Result<Player>> UpdateRole(string playerName, int newRoleId)
    {
        Result<Player> playerResult = await GetByNameAsync(playerName);
        if (!playerResult.Ok)
        {
            return playerResult.Error.Value;
        }
        
        var newRole = await _roleRepository.GetAsync(newRoleId);
        if (!newRole.Ok)
        {
            return newRole.Error.Value;
        }

        Player player = playerResult.Value;
        player.RoleId = newRoleId;
        player.Role = newRole.Value;
        var updatedPlayer = await _playerRepository.UpdateAsync(player);
        return updatedPlayer;
    }

    public async Task<Result<List<Player>>> GetAllAsync()
    {
        return await _playerRepository.GetAllAsync();
    }

    public async Task<Result<Player>> DeleteAsync(int id)
    {
        return await _playerRepository.DeletePlayerAsync(id);
    }

    public async Task<Result<List<Player>>> GetPlayersVisibleToPlayerAsync(string playerName)
    {
        var playerResult = await _playerRepository.GetByNameAsync(playerName);
        if (!playerResult.Ok)
        {
            return playerResult.Error.Value;
        }
        return await _playerRepository.GetPlayersVisibleToPlayerAsync(playerResult.Value);
    }

    public async Task<Result<List<Player>>> GetPlayersWithRoleAsync(int roleId)
    {
        var role = await _roleRepository.GetAsync(roleId);
        return role.Map(r => r.PlayersWithRole.ToList());
    }

    public async Task<Result<Role>> GetRoleFromPlayerAsync(string playerName)
    {
        return await _playerRepository.GetRoleFromPlayerAsync(playerName);
    }
}

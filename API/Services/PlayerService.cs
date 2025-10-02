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
        Player newPlayer = new()
        {
            Name = username
        };
        return await _playerRepository.CreateAsync(newPlayer);
    }

    public async Task<Result<Player>> UpdateRoleAsync(string playerName, int newRoleId)
    {
        var playerResult = await _playerRepository.GetByNameAsync(playerName);
        if (!playerResult.IsSuccess)
        {
            return playerResult;
        }

        var roleResult = await _roleRepository.GetAsync(newRoleId);
        if (!roleResult.IsSuccess)
        {
            return roleResult.Error;
        }

        Player player = playerResult.Value;
        Role role = roleResult.Value;
        player.RoleId = newRoleId;
        player.Role = role;
        return await _playerRepository.UpdateAsync(player);
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
        return await _playerRepository.GetByNameAsync(playerName)
            .ThenAsync(_playerRepository.GetPlayersVisibleToPlayerAsync);       
    }

    public async Task<Result<List<Player>>> GetPlayersWithRoleAsync(int roleId)
    {
        var repositoryResult = await _roleRepository.GetAsync(roleId);
        return repositoryResult.Map(static r => r.PlayersWithRole.ToList());
    }

    public async Task<Result<Role>> GetRoleFromPlayerAsync(string playerName)
    {
        return await _playerRepository.GetRoleFromPlayerAsync(playerName);
    }
}

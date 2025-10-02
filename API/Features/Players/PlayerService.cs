using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;

namespace API.Features.Players;

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
        var player = await _playerRepository.GetByIdAsync(id);
        if(player is null)
        {
            return Errors.ResourceNotFound("Player", id);
        }
        return player;
    }

    public async Task<Result<Player>> GetByNameAsync(string playerName)
    {
        var player = await _playerRepository.GetByNameAsync(playerName);
        if (player is null)
        {
            return Errors.ResourceNotFound($"Player with username {playerName} not found.");
        }
        return player;
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
        Player? player = await _playerRepository.GetByNameAsync(playerName);
        if (player is null)
        {
            return Errors.ResourceNotFound($"Player with username {playerName} not found.");
        }

        Role? role = await _roleRepository.GetByIdAsync(newRoleId);
        if (role is null)
        {
            return Errors.ResourceNotFound("Role", newRoleId);
        }
       
        player.RoleId = newRoleId;
        player.Role = role;
        await _playerRepository.UpdateAsync(player);
        return player;
    }

    public async Task<Result<List<Player>>> GetAllAsync()
    {
        return await _playerRepository.GetAllAsync();
    }

    public async Task<Result<bool>> DeleteAsync(int id)
    {
        Player? player = await _playerRepository.GetByIdAsync(id);
        if (player is null)
        {
            return Errors.ResourceNotFound("Player", id);
        }
        await _playerRepository.DeleteAsync(player);
        return true;
    }

    public async Task<Result<List<Player>>> GetPlayersVisibleToPlayerAsync(string playerName)
    {
        Player? player = await _playerRepository.GetByNameAsync(playerName);
        if(player is null)
        {
            return Errors.ResourceNotFound($"Player with username {playerName} not found.");
        }
        return await _playerRepository.GetPlayersVisibleToPlayerAsync(player);       
    }

    public async Task<Result<List<Player>>> GetPlayersWithRoleAsync(int roleId)
    {
        Role? role = await _roleRepository.GetByIdAsync(roleId);
        if(role is null)
        {
            return Errors.ResourceNotFound("Role", roleId);
        }
        return role.PlayersWithRole.ToList();
    }

    public async Task<Result<Role>> GetRoleFromPlayerAsync(string playerName)
    {
        Role? role = await _roleRepository.GetByPlayerNameAsync(playerName);
        if (role is null)
        {
            return Errors.ResourceNotFound($"Role of player {playerName} not found.");
        }
        return role;
    }
}

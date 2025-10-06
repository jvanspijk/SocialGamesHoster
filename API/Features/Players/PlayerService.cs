using API.Features.Players.Responses;
using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;

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

    public async Task<Result<PlayerResponse>> GetAsync(int id)
    {
        PlayerResponse? player = await _playerRepository.AsQueryable()
            .Where(p => p.Id == id)
            .Select(p => new PlayerResponse(p.Id, p.Name))
            .FirstOrDefaultAsync();

        if (player is null)
        {
            return Errors.ResourceNotFound("Player", id);
        }

        return player;
    }

    public async Task<Result<PlayerResponse>> GetByNameAsync(string playerName)
    {
        PlayerResponse? player = await _playerRepository.AsQueryable()
            .Where(p => p.Name == playerName)
            .Select(PlayerResponse.Projection)
            .FirstOrDefaultAsync();

        if (player is null)
        {
            return Errors.ResourceNotFound($"Player with username {playerName} not found.");
        }
        return player;
    }

    public async Task<Result<List<PlayerResponse>>> GetAllAsync()
    {
        List<PlayerResponse> players = await _playerRepository.AsQueryable()
            .Select(PlayerResponse.Projection)
            .ToListAsync();
        return players;
    }     

    public async Task<Result<List<PlayerResponse>>> GetPlayersBelongingToRoleAsync(int roleId)
    {
        Role? role = await _roleRepository.AsQueryable().Where(r => r.Id == roleId)
            .Include(r => r.PlayersWithRole)
            .FirstOrDefaultAsync();

        if (role is null)
        {
            return Errors.ResourceNotFound("Role", roleId);
        }

        return role.PlayersWithRole.Select(p => new PlayerResponse(p.Id, p.Name)).ToList();
    }

    public async Task<Result<FullPlayerResponse>> GetFullPlayer(int id)
    {
        var roleQuery = _playerRepository.AsQueryable()
                                     .Select(p => p.Role);
        var abilityQuery = roleQuery
            .SelectMany(r => r != null ? r.AbilityAssociations : new List<AbilityResponse>())
            .Select(ra => ra.Ability);

        FullPlayerResponse? player = await _playerRepository.AsQueryable()
            .Where(p => p.Id == id)
            .Select(p => new FullPlayerResponse(
                p.Id,
                p.Name,
                p.Role != null ? new RoleResponse(p.Role.Id, p.Role.Name, p.Role.Description, p.Role.AbilityAssociations) : null,
                p.CanSee.Select(ps => new PlayerResponse(ps.Id, ps.Name)).ToList(),
                p.CanBeSeenBy.Select(ps => new PlayerResponse(ps.Id, ps.Name)).ToList()
            )).FirstOrDefaultAsync();
    }

    public async Task<Result<bool>> CanPlayerSeeAsync(string sourcePlayerName, string targetPlayerName)
    {
        Player? source = await _playerRepository.GetByNameAsync(sourcePlayerName);
        if (source is null)
        {
            return Errors.ResourceNotFound($"Source player with username {sourcePlayerName} not found.");
        }

        Player? target = await _playerRepository.GetByNameAsync(targetPlayerName);
        if (target is null)
        {
            return Errors.ResourceNotFound($"Target player with username {targetPlayerName} not found.");
        }

        bool canSee = await _playerRepository.IsVisibleToPlayerAsync(source, target);
        return canSee;
    }
}

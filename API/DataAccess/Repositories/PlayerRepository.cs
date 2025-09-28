using API.Models;
using API.Validation;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class PlayerRepository
{
    private readonly APIDatabaseContext _context;
    public PlayerRepository(APIDatabaseContext context)
    {
        _context = context;
    }

    public async Task<Result<Player>> GetPlayerByIdAsync(int id)
    { 
        Player? player = await _context.Players.SingleOrDefaultAsync(
            p => p.Id == id
        );
        if (player == null)
        {
            return Errors.ResourceNotFound("Player", id.ToString());
        }
        return player;
    }

    public async Task<Result<Player>> GetPlayerByNameAsync(string name)
    {        
        Player? player = await _context.Players
            .SingleOrDefaultAsync(p => p.Name == name);

        if (player == null)
        {
            return Errors.ResourceNotFound("Player", name.ToString());

        }
        return player;       
    }

    public async Task<Result<Role>> GetRoleFromPlayerAsync(string playerName)
    {        
        Role? role = await _context.Players
            .Where(p => p.Name == playerName && p.Role != null)
            .Select(p => p.Role!)
            .Include(r => r.AbilityAssociations)
                .ThenInclude(ra => ra.Ability)
            .Include(r => r.CanSee)
            .Include(r => r.CanBeSeenBy)
            .SingleOrDefaultAsync();

        if (role == null)
        {
            return Errors.ResourceNotFound($"Could not find role for {playerName}.");
        }

        return role;
    }

    public async Task<Result<List<Player>>> GetPlayersAsync()
    {
        List<Player> players = await _context.Players.Include(p => p.Role).ToListAsync() ?? [];
        return players;
    }

    public async Task<Result<Player>> AddPlayerAsync(string name)
    {
        _context.Players.Add(new Player
        {
            Name = name
        });       
        await _context.SaveChangesAsync();
        return await GetPlayerByNameAsync(name);
    }

    public async Task<Result<Player>> UpdateRole(string playerName, int newRoleId)
    {
        Result<Player> playerResult = await GetPlayerByNameAsync(playerName);
        if (!playerResult.Ok)
        {
            return playerResult.Error.Value;
        }

        Player player = playerResult.Value;
        Role? newRole = await _context.Roles.FindAsync(newRoleId);
        if (newRole == null)
        {
            return Errors.ResourceNotFound("Role", newRoleId.ToString());
        }

        player.RoleId = newRoleId;
        player.Role = newRole;
        await _context.SaveChangesAsync();        
        return player;
    }

    public async Task<Result<Player>> DeletePlayerAsync(int id)
    {
        Result<Player> playerResult = await GetPlayerByIdAsync(id);
        if (!playerResult.Ok)
        {
            return Errors.ResourceNotFound("Player", id.ToString());
        }
        
        Player player = playerResult.Value;
        _context.Players.Remove(player);
        await _context.SaveChangesAsync();
        
        return player;
    }

    public async Task<Result<bool>> IsVisibleToPlayer(string playerName, string targetName)
    {
        var playerResult = await GetPlayerByNameAsync(playerName);
        if (!playerResult.Ok)
        {
            return playerResult.Error.Value;
        }

        Player player = playerResult.Value!;
        int? playerRoleId = player.RoleId;

        bool isVisible = await _context.Players
            .Where(p => p.Name == targetName)
            .Where(p =>
                p.Name == playerName ||
                p.CanBeSeenBy.Any(v => v.Player.Name == playerName) ||
                (p.Role != null &&
                 p.Role.CanBeSeenBy.Any(rv => rv.RoleId == playerRoleId))
            )
            .AnyAsync();

        return isVisible;
    }

    public async Task<Result<List<Player>>> GetPlayersVisibleToPlayerAsync(string playerName)
    {
        Result<Player> playerResult = await GetPlayerByNameAsync(playerName);
        if (!playerResult.Ok)
        {
            return playerResult.Error.Value;
        }

        Player player = playerResult.Value;

        IQueryable<Player> query =
            from p in _context.Players.Include(p => p.Role)
            where
                p.Id == player.Id ||
                p.CanBeSeenBy.Any(v => v.PlayerId == player.Id) ||
                (p.Role != null &&
                 p.Role.CanBeSeenBy.Any(rv => rv.RoleId == player.RoleId))
            select p;

        return await query.ToListAsync();
    }

}
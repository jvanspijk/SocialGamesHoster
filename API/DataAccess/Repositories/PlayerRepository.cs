using API;
using API.Models;
using API.Validation;
using LanguageExt.Common;
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
        try
        {
            Player? player = await _context.Players.SingleOrDefaultAsync(
                p => p.Id == id
            );
            if (player == null)
            {
                return new Result<Player>(
                    new NotFoundException($"Player with id '{id}' not found.")
                );
            }
            return player;
        }
        catch (Exception ex)
        {
            //TODO:
            //Log the exception, and add exception type to the function to handle
            //it explicitely
            return new Result<Player>(ex);
        }
    }

    public async Task<Result<Player>> GetPlayerByNameAsync(string name)
    {
        try
        {
            Player? player = await _context.Players
                .SingleOrDefaultAsync(p => p.Name == name);

            if (player == null)
            {
                return new Result<Player>(
                    new NotFoundException($"Player with name '{name}' not found.")
                );
            }
            return player;
        }
        catch (Exception ex)
        {
            //TODO:
            //Log the exception, and add exception type to the function to handle
            //it explicitely
            return new Result<Player>(ex);
        }
    }

    public async Task<Result<Role>> GetRoleFromPlayerAsync(string playerName)
    {
        try
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
                return new Result<Role>(new NotFoundException($"Role for player '{playerName}' not found."));
            }

            return role;
        }
        catch (Exception ex)
        {
            return new Result<Role>(ex);
        }
    }

    public async Task<Result<IEnumerable<Player>>> GetPlayersAsync()
    {
        return await _context.Players.Include(p => p.Role).ToListAsync();
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
        Console.WriteLine($"123Updating role for player {playerName} to role ID {newRoleId}");
        Result<Player> playerResult = await GetPlayerByNameAsync(playerName);
        if (!playerResult.IsSuccess)
        {
            return playerResult;
        }
        Player player = playerResult.GetValueOrThrow();
        Role? newRole = await _context.Roles.FindAsync(newRoleId);
        if (newRole == null)
        {
            return new Result<Player>(new NotFoundException($"Role with id {newRoleId} not found."));
        }
        player.RoleId = newRoleId;
        player.Role = newRole;
        await _context.SaveChangesAsync();        
        return player;
    }

    public async Task<Result<Player>> DeletePlayerAsync(int id)
    {
        Result<Player> playerResult = await GetPlayerByIdAsync(id);
        if (!playerResult.IsSuccess)
        {
            return playerResult;
        }
        
        Player player = playerResult.GetValueOrThrow();
        _context.Players.Remove(player);
        await _context.SaveChangesAsync();
        
        return player;
    }

    public async Task<Result<bool>> IsVisibleToPlayer(string playerName, string targetName)
    {
        var playerResult = await GetPlayerByNameAsync(playerName);
        if (!playerResult.IsSuccess)
        {
            return new Result<bool>(new NotFoundException($"{playerName} not found."));
        }

        Player player = playerResult.GetValueOrThrow();
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
        if (!playerResult.IsSuccess)
        {
            return new Result<List<Player>>(new NotFoundException($"{playerName} not found."));
        }

        Player player = playerResult.GetValueOrThrow();

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
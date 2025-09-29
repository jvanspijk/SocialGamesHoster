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

    public async Task<Result<Player>> GetAsync(int id)
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

    public async Task<Result<Player>> GetByNameAsync(string name)
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

    public async Task<Result<List<Player>>> GetAllAsync()
    {
        List<Player> players = await _context.Players.Include(p => p.Role).ToListAsync() ?? [];
        return players;
    }

    public async Task<Result<Player>> CreateAsync(Player player)
    {
        _context.Players.Add(player);       
        await _context.SaveChangesAsync();
        return await GetByNameAsync(player.Name);
    }

    public async Task<Result<Player>> UpdateAsync(Player updatedPlayer)
    {
        Result<Player> playerResult = await GetAsync(updatedPlayer.Id);
        if (!playerResult.Ok)
        {
            return playerResult.Error.Value;
        }
        Player player = playerResult.Value;
        player.Name = updatedPlayer.Name;
        player.RoleId = updatedPlayer.RoleId;
        await _context.SaveChangesAsync();        
        return player;
    }   

    public async Task<Result<Player>> DeletePlayerAsync(int id)
    {
        Result<Player> playerResult = await GetAsync(id);
        if (!playerResult.Ok)
        {
            return Errors.ResourceNotFound("Player", id.ToString());
        }
        
        Player player = playerResult.Value;
        _context.Players.Remove(player);
        await _context.SaveChangesAsync();
        
        return player;
    }

    public async Task<Result<bool>> IsVisibleToPlayer(Player source, Player target)
    {       
        int? playerRoleId = source.RoleId;

        return await _context.Players
            .Where(p => p.Name == target.Name)
            .Where(p =>
                p.Name == source.Name ||
                p.CanBeSeenBy.Any(v => v.Player.Name == source.Name) ||
                (p.Role != null &&
                 p.Role.CanBeSeenBy.Any(rv => rv.RoleId == playerRoleId))
            )
            .AnyAsync();
    }

    public async Task<Result<List<Player>>> GetPlayersVisibleToPlayerAsync(Player player)
    {
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
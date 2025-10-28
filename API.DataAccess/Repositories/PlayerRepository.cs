using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class PlayerRepository(APIDatabaseContext context) : IRepository<Player>
{
    private readonly APIDatabaseContext _context = context;

    #region Create
    public async Task<Player> CreateAsync(Player player)
    {
        _context.Players.Add(player);
        await _context.SaveChangesAsync();
        return player;
    }

    public async Task<List<Player>> CreateMultipleAsync(List<Player> players)
    {
        _context.Players.AddRange(players);
        await _context.SaveChangesAsync();
        return players;
    }
    #endregion

    #region Read
    public async Task<TProjectable?> GetAsync<TProjectable>(int id)
        where TProjectable : class, IProjectable<Player, TProjectable>
    {
        return await _context.Players
            .Where(a => a.Id == id)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }
    // TODO: check if game exists and return result object
    public async Task<TProjectable?> GetByNameAsync<TProjectable>(string name, int gameId) 
        where TProjectable : class, IProjectable<Player, TProjectable>
    {
        return await _context.Players
            .Where(p => p.Name == name)
            .Where(p => p.GameId == gameId)
            .Select(TProjectable.Projection)
            .FirstOrDefaultAsync();
    }

    public async Task<List<TProjectable>> GetAllAsync<TProjectable>() 
        where TProjectable : class, IProjectable<Player, TProjectable>
    {
        return await _context.Players            
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    // TODO: Should return result object to indicate if game exists
    public async Task<List<TProjectable>> GetAllFromGameAsync<TProjectable>(int gameId) 
        where TProjectable : class, IProjectable<Player, TProjectable>
    {
        return await _context.Players
            .Where(p => p.GameId == gameId)
            .Select(TProjectable.Projection)
            .ToListAsync();
    }

    public async Task<Result<List<TProjectable>>> GetMultipleAsync<TProjectable>(List<int> ids)
    where TProjectable : class, IProjectable<Player, TProjectable>
    {
        var foundPlayers = await _context.Players
            .Where(r => ids.Contains(r.Id))
            .Select(TProjectable.Projection)
            .ToListAsync();

        if (ids.Count != foundPlayers.Count)
        {
            var foundIds = await _context.Players
                .Where(r => ids.Contains(r.Id))
                .Select(r => r.Id)
                .ToHashSetAsync();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Players", "Ids", string.Join(", ", missingIds));
        }

        return foundPlayers;
    }

    public async Task<Player?> GetAsync(int id)
    {
        return await _context.Players.FindAsync(id);
    }

    public async Task<Result<List<Player>>> GetMultipleAsync(List<int> ids)
    {
        var foundPlayers = await _context.Players
            .Where(p => ids.Contains(p.Id))
            .ToListAsync();

        if(ids.Count != foundPlayers.Count)
        {
            var foundIds = foundPlayers.Select(p => p.Id).ToHashSet();
            var missingIds = ids.Where(id => !foundIds.Contains(id));
            return Errors.ResourceNotFound("Player", "Ids", string.Join(", ", missingIds));
        }

        return foundPlayers;
    }

    public async Task<Player?> GetByNameAsync(string name, int gameId)
    {
        return await _context.Players
            .Where(p => p.GameId == gameId)
            .Where(p => p.Name == name)
            .FirstOrDefaultAsync();
    }

    public async Task<bool> IsVisibleToPlayerAsync(Player source, Player target)
    {
        int? playerRoleId = source.RoleId;

        Player sourcePlayer = await _context.Players
            .Include(p => p.Role)
            .Include(p => p.CanSee)
            .SingleOrDefaultAsync(p => p.Id == source.Id)
            ?? throw new ArgumentException($"Source player with Id {source.Id} not found.");

        Player targetPlayer = await _context.Players
            .Include(p => p.Role)
            .Include(p => p.CanBeSeenBy)
            .SingleOrDefaultAsync(p => p.Id == target.Id)
            ?? throw new ArgumentException($"Target player with Id {target.Id} not found.");

        return sourcePlayer.CanSee.Contains(target);
    }

    public async Task<List<Player>> GetPlayersVisibleToPlayerAsync(Player player)
    {
        Player sourcePlayer = await _context.Players
           .Include(p => p.CanSee)
           .SingleAsync(p => p.Id == player.Id);       

        return [.. sourcePlayer.CanSee];
    }
    #endregion

    #region Update
    public async Task<Player> UpdateAsync(Player updatedPlayer)
    {
        _context.Entry(updatedPlayer).State = EntityState.Modified;
        _context.Players.Update(updatedPlayer);
        await _context.SaveChangesAsync();
        return updatedPlayer;
    }
    #endregion

    #region Delete
    public async Task DeleteAsync(Player player)
    {
        _context.Entry(player).State = EntityState.Modified;
        _context.Players.Remove(player);
        await _context.SaveChangesAsync();
    }
    #endregion
}
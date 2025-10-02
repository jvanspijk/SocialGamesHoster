using API.Domain.Models;
using Microsoft.EntityFrameworkCore;

namespace API.DataAccess.Repositories;

public class PlayerRepository : IRepository<Player>
{
    private readonly APIDatabaseContext _context;
    public PlayerRepository(APIDatabaseContext context)
    {
        _context = context;
    }
    // Create
    public async Task<Player> CreateAsync(Player player)
    {
        _context.Players.Add(player);
        await _context.SaveChangesAsync();
        return player;
    }
    // Read
    public async Task<Player?> GetByIdAsync(int id)
    { 
        return await _context.Players.SingleOrDefaultAsync(
            p => p.Id == id
        );       
    }
    public async Task<Player?> GetByNameAsync(string name)
    {        
        return await _context.Players
            .SingleOrDefaultAsync(p => p.Name == name);           
    }

    public async Task<List<Player>> GetAllAsync()
    {
        return await _context.Players.Include(p => p.Role).ToListAsync();
    }

    public async Task<bool> IsVisibleToPlayerAsync(Player source, Player target)
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

    public async Task<List<Player>> GetPlayersVisibleToPlayerAsync(Player player)
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

    // Update
    public async Task UpdateAsync(Player updatedPlayer)
    {
        _context.Entry(updatedPlayer).State = EntityState.Modified;
        _context.Players.Update(updatedPlayer);
        await _context.SaveChangesAsync();
    }   

    // Delete
    public async Task DeleteAsync(Player player)
    {
        _context.Entry(player).State = EntityState.Modified;
        _context.Players.Remove(player);
        await _context.SaveChangesAsync();
    }  
    

}
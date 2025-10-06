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

    public IQueryable<Player> AsQueryable() => _context.Players.AsNoTracking();

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
        return await _context.Players
            .Include(p => p.Role)
            .Include(p => p.CanSee)
            .Include(p => p.CanBeSeenBy)
            .SingleOrDefaultAsync(p => p.Id == id);       
    }
    public async Task<Player?> GetByNameAsync(string name)
    {
        return await _context.Players
            .Include(p => p.Role)
                .ThenInclude(r => r.Abilities)
            .Include(p => p.CanSee)
            .Include(p => p.CanBeSeenBy)
            .SingleOrDefaultAsync(p => p.Name == name);           
    }

    public async Task<List<Player>> GetAllAsync()
    {
        return await _context.Players.Include(p => p.Role).ToListAsync();
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

        return await _context.Players
            .Where(p => p.Name == target.Name)
            .Where(p =>
                p.Name == source.Name ||
                p.CanBeSeenBy.Any(p => p.Name == source.Name) ||
                (p.Role != null &&
                 p.Role.CanBeSeenBy.Any(r => r.Id == playerRoleId))
            )
            .AnyAsync();
    }

    public async Task<List<Player>> GetPlayersVisibleToPlayerAsync(Player player)
    {
        IQueryable<Player> query =
            from p in _context.Players.Include(p => p.Role)
            where
                p.Id == player.Id ||
                p.CanBeSeenBy.Any(p => p.Id == player.Id) ||
                (p.Role != null &&
                 p.Role.CanBeSeenBy.Any(r => r.Id == player.RoleId))
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
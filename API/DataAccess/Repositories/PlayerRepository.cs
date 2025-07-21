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

    public async Task<Result<Player>> GetPlayerAsync(string name)
    {
        try
        {
            Player? player = await _context.Players.SingleAsync(
                p => p.Name == name
            );
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
            return new Result<Player>(ex);
        }        
    }

    public async Task<Result<IEnumerable<Player>>> GetPlayersAsync()
    {
        return await _context.Players.ToListAsync();
    }

    public async Task<Result<Player>> AddPlayerAsync(string name)
    {
        _context.Players.Add(new Player
        {
            Name = name
        });       
        await _context.SaveChangesAsync();
        return await GetPlayerAsync(name);
    }
}
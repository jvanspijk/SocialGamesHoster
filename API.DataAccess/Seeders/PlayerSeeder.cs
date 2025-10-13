using API.Domain.Models;
using Microsoft.EntityFrameworkCore;
using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace API.DataAccess.Seeders;

internal class PlayerSeeder
{
    public List<Player> Players = [];
    public PlayerSeeder SeedPlayers(int gameId)
    {       

        List<string> testUserNames = [
           "Alice", "John", "Emily", "Michael", "Sarah",
           "Jessica", "David", "Ashley", "Matthew", "Amanda",
           "Joshua", "Jennifer", "Daniel", "Elizabeth", "James",
           "Charlie", "Kyle", "Bob", "Megan", "Laura",
        ];

        List<Player> players = new(testUserNames.Count);
        foreach (var userName in testUserNames)
        {
            players.Add(new Player
            {
                Name = userName,
                GameId = gameId
            });
        }

        Players = players;

        return this;
    }

    public PlayerSeeder AddRoles(List<Role> roles)
    {
        if (Players.Count == 0 || roles.Count == 0)
        {
            throw new InvalidOperationException("Players or roles is empty.");
        }

        for (int i = 0; i < Players.Count; i++)
        {
            int roleId = i % roles.Count + 1; // Cycle through roles

            // Append role name to player name for easy debugging
            string name = Players[i].Name + " " + roles[i % roles.Count].Name;

            Players[i].Id = roleId;
            Players[i].RoleId = roleId;
            Players[i].Name = name;            
        }

        return this;
    }    

    public void ApplyTo(ModelBuilder builder)
    {
        if (Players.Count == 0)
        {
            throw new InvalidOperationException("No players to apply. Did you forget to call SeedPlayers()?");
        }
        builder.Entity<Player>().HasData(Players);
    }
}

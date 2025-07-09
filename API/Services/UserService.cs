using API.Models;
using System.Collections.Concurrent;

namespace API.Services;

public class UserService
{
    // For production, this needs to be a concurrent dictionary
    private readonly List<User> _users = [];
    public UserService()
    {
        Random rnd = new Random();
        List<Role> roles = [
            new Role { Name = "Rol 1", Description = "Je kan eigenlijk niet zo veel." },
            new Role { Name = "Rol 2", Description = "Je kan nog minder." },
            new Role { Name = "Rol 3", Description = "Heel nuttig ben je niet."},
        ];

        List<string> testUserNames = ["Bob", "Jan", "Kees", "Piet", "Jaap"];
        foreach (string name in testUserNames)
        {
            int randomIndex = rnd.Next(roles.Count);
            Role assignedRole = roles[randomIndex];
            _users.Add(new User { Name = name, Role = assignedRole });
        }
    }

    public User? GetUser(string name)
    {
        return _users.FirstOrDefault(
            u => User.NormalizeName(u.Name) == User.NormalizeName(name)
        );
    }

    public List<User> GetUsers()
    {
        return _users.ToList();
    }

    public bool UserIsOnline(string username)
    {
        var user = GetUser(username);
        return user?.IsOnline ?? false;
    }

    public List<User> GetOnlineUsers()
    {
        return _users.Where(u => u.IsOnline).ToList();
    }

    public User LogIn(string username)
    {
        var user = GetUser(username);
        if (user == null)
        {
            throw new UserDoesNotExistException($"Can't find {username}.");
        }
        user.LogIn();
        return user;
    }

    public void LogOut(string username)
    {
        var user = GetUser(username);
        if (user == null)
        {
            throw new UserDoesNotExistException($"Can't find {username}.");
        }
        user.LogOut();
    }

    public User AddUser(string name)
    {
        // Protect invariants.
        throw new NotImplementedException();
    }
}
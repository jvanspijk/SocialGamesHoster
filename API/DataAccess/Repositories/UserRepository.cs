using API.Models;
using API.Services;
using API.Validation;
using LanguageExt.Common;

namespace API.DataAccess.Repositories;

public class UserRepository
{
    // For production, this needs to be a concurrent dictionary
    private readonly List<User> _users = [];
    public UserRepository()
    {
        List<Role> roles = [
            new Role { Id = 1, Name = "Rol 1", Description = "Je kan eigenlijk niet zo veel." },
            new Role { Id = 2, Name = "Rol 2", Description = "Je kan nog minder." },
            new Role { Id = 3, Name = "Rol 3", Description = "Heel nuttig ben je niet."},
        ];

        List<string> testUserNames = ["Bob", "Jan", "Kees", "Piet", "Jaap"];

        Random rnd = new();

        foreach (string name in testUserNames)
        {
            int randomIndex = rnd.Next(roles.Count);
            Role assignedRole = roles[randomIndex];
            _users.Add(new User { Name = name, Role = assignedRole });
        }

        // We pick out a random user to login
        // This is for testing purposes and should not
        // be used in production since it can cause
        // a deadlock.
        Task.Run(() => LogInAsync("Kees")).Wait();
    }

    public async Task<Result<User>> GetUserAsync(string name)
    {
        try
        {
            return _users.First(
                u => User.NormalizeName(u.Name) == User.NormalizeName(name)
            );
        }
        catch (Exception ex)
        {
            return new Result<User>(ex);
        }        
    }

    public async Task<Result<IEnumerable<User>>> GetUsersAsync()
    {
        return _users.ToList();
    }

    public async Task<bool> UserIsOnline(string username)
    {
        Result<User> userResult = await GetUserAsync(username);

        return userResult.Match(
            user =>
            {
               return user.IsOnline; 
            }, 
            exception =>
            {
                return false;
            }
        );
    }

    public async Task<Result<User>> LogInAsync(string username)
    {
        Result<User> userResult = await GetUserAsync(username);

        return userResult.Match(
            user =>
            {
                user.LogIn();
                return user;
            },
            exception =>
            {
                return new Result<User>(exception);
            }
        );
    }

    public async Task<Result<User>> LogOutAsync(string username, string sessionId)
    {
        var userResult = await GetUserAsync(username);
        return userResult.Match(
            user =>
            {
                user.LogOut();
                return user;
            },
            exception =>
            {
                return new Result<User>(exception);
            }
        );
    }

    public async Task<Result<User>> AddUserAsync(string name)
    {
        // Protect invariants. <- I meant weird names
        return new Result<User>(new NotImplementedException());
    }
}
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
        List<Role> roles = new(8);

        for (int i = 1; i <= 8; i++)
        {
            roles.Add(new Role { Id = i, Name = i.ToString() });
        }

        List<string> testUserNames = [            
            "Alice", "John", "Emily", "Michael", "Sarah",
            "Jessica", "David", "Ashley", "Matthew", "Amanda",
            "Joshua", "Jennifer", "Daniel", "Elizabeth", "James"
        ];

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
        Task.Run(() => LogInAsync("James")).Wait();
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
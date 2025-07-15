
using Microsoft.AspNetCore.Identity;
using System.Text.Json.Serialization;
using API.Validation;

namespace API.Models;

public class User
{
    public required string Name { get; set; }
    public bool IsOnline { get; private set; } = false;

    [JsonIgnore]
    public Role? Role { get; set; }
    public static string NormalizeName(string userName)
    {
        return userName.ToLowerInvariant().Normalize();
    }

    public void LogIn()
    {
        if (IsOnline)
        {
            throw new UserAlreadyLoggedInException($"User {Name} is already logged in.");
        }
        IsOnline = true;
        Console.WriteLine($"User '{Name}' logged in.");
    }

    public void LogOut()
    {
        if (!IsOnline)
        {
            throw new UserNotOnlineException($"User {Name} is not online and cannot log out.");
        }
        IsOnline = false;
        Console.WriteLine($"User {Name} logged out.");
    }
}



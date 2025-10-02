using System.ComponentModel.DataAnnotations;

namespace API.Features.Authentication.Requests;

public readonly record struct AdminLoginRequest(string Username, string Password);


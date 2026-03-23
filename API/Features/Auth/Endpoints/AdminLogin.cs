using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Auth.Endpoints;

public static class AdminLogin
{
    public readonly record struct Request(string Username, string PasswordHash);
    public readonly record struct Response(string Token);
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(AuthService authService, Request request)
    {
        if (await authService.AdminCredentialsAreValid(request.Username, request.PasswordHash))
        {
            string token = authService.GenerateAdminToken();
            return APIResults.Ok(new Response{ Token = token });
        }
        return APIResults.Unauthorized();
    }
}

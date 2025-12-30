namespace API.Features.Auth.Endpoints;

public static class AdminLogin
{
    public readonly record struct Request(string Username, string PasswordHash);
    public readonly record struct Response(string Token);
    public static async Task<IResult> HandleAsync(AuthService authService, Request request)
    {
        if (await authService.AdminCredentialsAreValid(request.Username, request.PasswordHash))
        {
            string token = authService.GenerateAdminToken();
            return Results.Ok(new Response{ Token = token });
        }
        return Results.Unauthorized();
    }
}

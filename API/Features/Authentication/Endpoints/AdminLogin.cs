namespace API.Features.Authentication.Endpoints;

public static class AdminLogin
{
    public static async Task<IResult> HandleAsync(AuthService authService, string username, string passwordHash)
    {
        if (await authService.AdminCredentialsAreValid(username, passwordHash))
        {
            string token = authService.GenerateAdminToken(username);
            return Results.Ok(new { Token = token, Role = "admin" });
        }
        return Results.Unauthorized();
    }
}

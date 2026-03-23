using Microsoft.AspNetCore.Http.HttpResults;

namespace API.Features.Auth.Endpoints;

public static class PlayerLogout
{
    public readonly record struct Request(int PlayerId, string Token);
    public readonly record struct Response;
    public static async Task<Results<Ok, ProblemHttpResult>> HandleAsync()
    {
        // Check if the player id belongs to the token or if the token is from an admin
        return APIResults.Ok();
    }
}

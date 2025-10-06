using API.DataAccess.Repositories;
using API.Domain.Models;

namespace API.AdminFeatures.Players.Endpoints;
public static class CreatePlayer
{
    public readonly record struct Request(string Name);
    public readonly record struct Response(int Id, string Name);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request)
    {
        // TODO: validation
        Player player = new() { Name = request.Name };
        Player result = await repository.CreateAsync(player);
        return Results.Ok(new Response(result.Id, result.Name));
    }
}

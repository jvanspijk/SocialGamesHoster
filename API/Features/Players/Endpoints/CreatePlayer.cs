using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Players.Responses;

namespace API.Features.Players.Endpoints;
public static class CreatePlayer
{
    public readonly record struct Request(string Name);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request, int gameId, HttpContext context)
    {
        // TODO: validation
        Player player = new() { Name = request.Name, GameId = gameId };
        Player result = await repository.CreateAsync(player);

        PlayerNameResponse response = new(result.Id, result.Name);
        string uri = context.Request.Scheme + "://" +
            context.Request.Host.ToString() + 
            $"/players/{result.Id}";

        return Results.Created(uri, response);
    }
}

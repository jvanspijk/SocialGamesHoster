using API.DataAccess.Repositories;
using API.Domain.Models;

namespace API.AdminFeatures.Players.Endpoints;
public static class CreatePlayer
{
    public readonly record struct Request(string Name);
    public readonly record struct Response(int Id, string Name);
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request, HttpContext context)
    {
        // TODO: validation
        Player player = new() { Name = request.Name };
        Player result = await repository.CreateAsync(player);

        Response response = new(result.Id, result.Name);
        string uri = context.Request.Scheme + "://" +
            context.Request.Host.ToString() + 
            $"/players/{result.Id}";

        return Results.Created(uri, response);
    }
}

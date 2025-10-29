using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Domain.Validation;
using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;
public static class CreatePlayer
{
    public readonly record struct Request(string Name) : IValidatable<Request>
    {
        public IEnumerable<ValidationError> Validate()
        {
            if (Name.Length is < 1 or > 32)
            {
                yield return new ValidationError(nameof(Name), "Player name must be between 1 and 32 characters long.");
            }
        }
    }
    public record Response(int Id, string Name) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(player.Id, player.Name);
    }
    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request, int gameId)
    {
        var validationResult = request.Validate();
        if (validationResult.HasErrors())
        {
            return Results.ValidationProblem(validationResult.ToProblemDetails());
        }

        Player player = new() { Name = request.Name, GameId = gameId };
        Player result = await repository.CreateAsync(player);

        Response response = new(result.Id, result.Name);

        return Results.Ok(response);
    }
}

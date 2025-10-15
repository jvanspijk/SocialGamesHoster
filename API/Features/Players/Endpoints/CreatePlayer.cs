using API.DataAccess.Repositories;
using API.Domain;
using API.Domain.Models;
using API.Domain.Validation;
using API.Features.Players.Responses;
using System.ComponentModel.DataAnnotations;

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

    public static async Task<IResult> HandleAsync(PlayerRepository repository, Request request, int gameId, HttpContext context)
    {
        var validationResult = request.Validate();
        if (validationResult.HasErrors())
        {
            return Results.ValidationProblem(validationResult.ToProblemDetails());
        }

        Player player = new() { Name = request.Name, GameId = gameId };
        Player result = await repository.CreateAsync(player);

        PlayerNameResponse response = new(result.Id, result.Name);
        string uri = context.Request.Scheme + "://" +
            context.Request.Host.ToString() + 
            $"/players/{result.Id}";

        return Results.Created(uri, response);
    }
}

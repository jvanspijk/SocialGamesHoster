using API.DataAccess;
using API.Domain.Entities;
using Microsoft.AspNetCore.Http.HttpResults;
using System.ComponentModel.DataAnnotations;
using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;
public static class CreatePlayer
{
    public readonly record struct Request(string Name) : IValidatableObject
    {
        public IEnumerable<ValidationResult> Validate(ValidationContext validationContext)
        {
            if (Name.Length is < 1 or > 32)
            {
                yield return new ValidationResult("Player name must be between 1 and 32 characters long.", [nameof(Name)]);
            }
        }
    }
    public record Response(int Id, string Name) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(player.Id, player.Name);
    }
    public static async Task<Results<CreatedAtRoute<Response>, ProblemHttpResult>> HandleAsync(IRepository<Player> repository, Request request, int gameId)
    {        
        Player player = new() { Name = request.Name, GameId = gameId };
        repository.Add(player);
        await repository.SaveChangesAsync();

        Response response = new(player.Id, player.Name);        
        return APIResults.CreatedAtRoute(response, nameof(GetPlayer), player.Id);
    }
}

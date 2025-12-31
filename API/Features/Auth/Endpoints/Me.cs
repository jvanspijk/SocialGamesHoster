using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Auth.Common;
using System.Linq.Expressions;
using System.Security.Claims;

namespace API.Features.Auth.Endpoints;

public static class Me
{
    public record Response(int Id, string Name, RoleInfo? Role) : IProjectable<Player, Response>
    {
        public static Expression<Func<Player, Response>> Projection =>
            player => new Response(
                player.Id,
                player.Name,
                player.Role == null ? null :
                new RoleInfo(player.Role.Id, player.Role.Name, player.Role.Description,
                    player.Role.Abilities
                    .Select(a => new AbilityInfo(a.Id, a.Name, a.Description))
                    .ToList()
                )
            );
    }
    public static async Task<IResult> HandleAsync(HttpRequest request, PlayerRepository repository)
    {
        var playerClaimsResult = AuthService.GetPlayerClaims(request);
        if (playerClaimsResult.IsFailure)
        {
            return playerClaimsResult.AsIResult();
        }

        (int playerId, int? _) = playerClaimsResult.Value;

        var response = await repository.GetAsync<Response>(playerId);
        if(response == null)
        {
            return Results.NotFound($"Player {playerId} not found");
        }

        return Results.Ok(response);
    }
}

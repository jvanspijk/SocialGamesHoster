using API.DataAccess;
using API.Domain.Entities;
using API.Features.Auth.Common;
using Microsoft.AspNetCore.Http.HttpResults;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;
using System.Security.Claims;

namespace API.Features.Auth.Endpoints;

public static class Me
{
    private static string CacheKey(int playerId) => $"{nameof(Me)}_{playerId}";
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
    public static async Task<Results<Ok<Response>, ProblemHttpResult>> HandleAsync(
        IRepository<Player> repository,
        IMemoryCache cache,
        HttpRequest request)
    {
        var playerClaimsResult = AuthService.GetPlayerClaims(request);

        if (playerClaimsResult.IsFailure)
        {
            // TODO: return specific failure reason (e.g., token missing, token invalid, etc.)
            return APIResults.Unauthorized();
        }

        var (playerId, _) = playerClaimsResult.Value;

        var response = await cache.GetOrCreateAsync(CacheKey(playerId), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromHours(1);
            return await repository.GetReadOnlyAsync<Response>(p => p.Id == playerId);
        });

        if (response == null)
        {
            return APIResults.NotFound<Player>(playerId);
        }

        return TypedResults.Ok(response);
    }
}

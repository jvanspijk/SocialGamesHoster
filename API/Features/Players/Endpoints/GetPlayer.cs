using API.DataAccess;
using API.Domain.Entities;
using API.Features.Auth;
using API.Features.Players.Common;
using Microsoft.Extensions.Caching.Memory;
using System.Linq.Expressions;

namespace API.Features.Players.Endpoints;

public static class GetPlayer
{
    private static string CacheKey(int playerId) => $"{nameof(GetPlayer)}_{playerId}";
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
    public static async Task<IResult> HandleAsync(IRepository<Player> repository, IMemoryCache cache, AuthService authService, HttpRequest request, int id)
    {
        var canSeeResult = await authService.CanSeePlayerAsync(request, id);
        if (canSeeResult.IsFailure)
        {
            return canSeeResult.AsIResult();
        }

        if (canSeeResult.Value == false)
        {
            return Results.Forbid();
        }

        Response? result = await cache.GetOrCreateAsync(CacheKey(id), async entry =>
        {
            entry.AbsoluteExpirationRelativeToNow = TimeSpan.FromMinutes(30);
            return await repository.GetReadOnlyAsync<Response>(p => p.Id == id);
        });

        if (result == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
        return Results.Ok(result);
    }
}

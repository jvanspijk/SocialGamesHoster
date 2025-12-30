using API.DataAccess;
using API.DataAccess.Repositories;
using API.Domain.Models;
using API.Features.Auth;
using API.Features.Players.Common;
using System.Linq.Expressions;
using System.Security.Claims;

namespace API.Features.Players.Endpoints;

public static class GetFullPlayer
{
    public record Response(int Id, string Name, RoleInfo? Role, 
        bool IsEliminated, List<int> CanSeeIds, List<int> CanBeSeenByIds) 
        : IProjectable<Player, Response>
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
                ),
                player.IsEliminated,
                player.CanSee.Select(p => p.Id).ToList(),
                player.CanBeSeenBy.Select(p => p.Id).ToList()
            );
    }
    public static async Task<IResult> HandleAsync(PlayerRepository repository, AuthService authService, ClaimsPrincipal claims, int id)
    {
        // there's no admin login on the front end so turn this off for now     
        //var authResult = authService.IsAdmin(claims);
        //if(authResult.IsFailure)
        //{
        //    return authResult.AsIResult();
        //}

        //bool isAdmin = authResult.Value;
        //if(!isAdmin)
        //{           
        //     //return Results.Forbid();       
        //}

        Response? result = await repository.GetAsync<Response>(id);
        if (result == null)
        {
            return Results.NotFound($"Player with id {id} not found.");
        }
        return Results.Ok(result);
    }
}

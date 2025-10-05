using API.Features.Abilities.Endpoints;
using API.Features.Abilities.Responses;
using API.Features.Players.Endpoints;
using API.Features.Players.Responses;
using API.Features.Roles.Endpoints;
using API.Features.Roles.Responses;
using API.Features.Rounds.Endpoints;
using API.Features.Rounds.Responses;

namespace API;

public static class GameEndpoints
{
    public static IEndpointRouteBuilder MapGameEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapAbilityEndpoints();
        builder.MapPlayerEndpoints();
        builder.MapRoleEndpoints();
        builder.MapRoundEndPoints();

        return builder;
    }

    private static RouteGroupBuilder MapAbilityEndpoints(this IEndpointRouteBuilder builder)
    {
        var abilityGroup = builder.MapGroup("/abilities");

        abilityGroup.MapGet("/", GetAbilities.HandleAsync)
            .WithName("GetAbilities")
            .Produces<List<AbilityResponse>>(StatusCodes.Status200OK);

        abilityGroup.MapGet("/{id:int}", GetAbility.HandleAsync)
            .WithName("GetAbility")
            .Produces<AbilityResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);


        return abilityGroup;
    }

    private static RouteGroupBuilder MapPlayerEndpoints(this IEndpointRouteBuilder builder)
    {
        var playerGroup = builder.MapGroup("/players")
            .WithTags("Players");

        playerGroup.MapGet("/", GetPlayers.HandleAsync)
            .WithName("GetAllPlayers")
            .Produces<List<PlayerResponse>>(StatusCodes.Status200OK); 

        playerGroup.MapGet("/{id:int}", GetPlayer.HandleByIdAsync)
            .WithName("GetPlayerById")
            .Produces<PlayerResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playerGroup.MapGet("/{name:alpha}", GetPlayer.HandleByNameAsync)
            .WithName("GetPlayerByName")
            .Produces<PlayerResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return playerGroup;
    }

    private static RouteGroupBuilder MapRoleEndpoints(this IEndpointRouteBuilder builder)
    {
        var roleGroup = builder.MapGroup("/roles")
            .WithTags("Roles");

        roleGroup.MapGet("/", GetRoles.HandleAsync)
            .WithName("GetAllRoles")
            .Produces<List<RoleResponse>>(StatusCodes.Status200OK);

        roleGroup.MapGet("/{id:int}", GetRole.HandleAsync)
            .WithName("GetRole")
            .Produces<RoleResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return roleGroup;
    }

    private static RouteGroupBuilder MapRoundEndPoints(this IEndpointRouteBuilder builder)
    {
        var roundGroup = builder.MapGroup("/current-round")
            .WithTags("CurrentRound");

        roundGroup.MapGet("/end-time", GetCurrentEndTime.HandleAsync)
            .WithName("GetCurrentEndTime")
            .Produces<RoundEndTimeResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return roundGroup;
    }
}

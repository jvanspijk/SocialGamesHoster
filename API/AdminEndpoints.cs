using API.AdminFeatures.Abilities.Endpoints;
using API.AdminFeatures.Abilities.Responses;
using API.AdminFeatures.Players.Endpoints;
using API.AdminFeatures.Players.Responses;
using API.AdminFeatures.Roles.Endpoints;
using API.AdminFeatures.Roles.Responses;

namespace API;

public static class AdminEndpoints
{
    public static IEndpointRouteBuilder MapAdminEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapAbilityEndpoints();
        builder.MapPlayerEndpoints();
        return builder;
    }

    private static RouteGroupBuilder MapAbilityEndpoints(this IEndpointRouteBuilder builder)
    {
        var abilityGroup = builder.MapGroup("/abilities")
            .WithTags("Abilities");

        abilityGroup.MapGet("/", GetAbilities.HandleAsync);

        abilityGroup.MapGet("/{id:int}", GetAbility.HandleAsync)
            .WithName("GetAbilityById")
            .Produces<AbilityResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return abilityGroup;
    }

    private static RouteGroupBuilder MapRoleEndpoints(this IEndpointRouteBuilder builder)
    {
        var abilityGroup = builder.MapGroup("/roles")
            .WithTags("Roles");

        // TODO:
        //abilityGroup.MapGet("/", GetRoles.HandleAsync);

        //abilityGroup.MapGet("/{id:int}", GetRole.HandleAsync)
        //    .WithName("GetAbilityById")
        //    .Produces<RoleResponse>(StatusCodes.Status200OK)
        //    .ProducesProblem(StatusCodes.Status404NotFound);

        return abilityGroup;
    }

    private static RouteGroupBuilder MapPlayerEndpoints(this IEndpointRouteBuilder builder)
    {
        var playerGroup = builder.MapGroup("/players")
            .WithTags("Players");

        playerGroup.MapGet("/", GetPlayers.HandleAsync);

        playerGroup.MapGet("/{id:int}", GetPlayer.HandleByIdAsync)
            .WithName("GetPlayerById")
            .Produces<FullPlayerResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playerGroup.MapGet("/{name:alpha}", GetPlayer.HandleByNameAsync)
            .WithName("GetPlayerByName")
            .Produces<FullPlayerResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return playerGroup;
    }
}

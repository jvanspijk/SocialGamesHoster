using API.AdminFeatures.Players.Endpoints;
using API.AdminFeatures.Players.Responses;

namespace API;

public static class AdminEndpoints
{
    public static IEndpointRouteBuilder MapAdminEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapPlayerEndpoints();
        return builder;
    }

    private static RouteGroupBuilder MapPlayerEndpoints(this IEndpointRouteBuilder builder)
    {
        var playerGroup = builder.MapGroup("/players")
            .WithTags("Players");

        playerGroup.MapGet("/", GetPlayers.HandleAsync);

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
}

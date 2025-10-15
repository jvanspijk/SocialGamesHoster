using API.Features.Abilities.Endpoints;
using API.Features.Abilities.Responses;
using API.Features.Players.Endpoints;
using API.Features.Players.Responses;
using API.Features.Roles.Endpoints;
using API.Features.Roles.Responses;
using API.Features.Rounds.Endpoints;
using API.Features.Rounds.Responses;
using API.Features.Rulesets.Endpoints;
using API.Features.Rulesets.Responses;

namespace API;

public static class Endpoints
{
    public static IEndpointRouteBuilder MapEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapRoleEndpoints();
        builder.MapAbilityEndpoints();
        builder.MapPlayerEndpoints();

        builder.MapGameEndpoints();
        builder.MapRulesetEndpoints();

        return builder;
    }

    private static RouteGroupBuilder MapRulesetEndpoints(this IEndpointRouteBuilder builder)
    {
        var rulesetsGroup = builder.MapGroup("/rulesets/{rulesetId:int}")
            .WithTags("Ruleset");

        rulesetsGroup.MapGet("/", GetRuleset.HandleAsync)
            .WithName("GetRulesetById")
            .Produces<RulesetResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var abilitiesGroup = rulesetsGroup.MapGroup("abilties")
            .WithTags("Ability");

        abilitiesGroup.MapPost("/", CreateAbility.HandleAsync)
            .WithName("CreateAbility");

        abilitiesGroup.MapGet("/", GetAbilitiesFromRuleset.HandleAsync)
            .WithName("GetRulesetAbilities")
            .Produces<List<AbilityResponse>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var rolesGroup = rulesetsGroup.MapGroup("/roles")
            .WithTags("Role");

        rolesGroup.MapPost("/", CreateRole.HandleAsync)
            .WithName("CreateRole");

        rolesGroup.MapGet("/", GetRoles.HandleAsync)
            .WithName("GetRulesetRoles")
            .Produces<List<RoleResponse>>(StatusCodes.Status200OK);

        return rulesetsGroup;
    }

    private static RouteGroupBuilder MapGameEndpoints(this IEndpointRouteBuilder builder)
    {
        var gamesGroup = builder.MapGroup("/games/{gameId:int}")
           .WithTags("GameSession");

        var playersGroup = gamesGroup.MapGroup("/players")
            .WithTags("Player");

        playersGroup.MapGet("/", GetPlayersFromGame.HandleAsync)
            .WithName("GetGamePlayers")
            .Produces<List<PlayerNameResponse>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapGet("/{name:alpha}", GetPlayerFromGame.HandleAsync)
            .WithName("GetPlayerByName")
            .Produces<PlayerNameResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapPost("/", CreatePlayer.HandleAsync)
            .WithName("CreatePlayer");

        var roundsGroup = gamesGroup.MapGroup("rounds")
            .WithTags("Round");

        roundsGroup.MapGet("/current", GetCurrentRound.HandleAsync)
            .WithName("GetCurrentRound");

        roundsGroup.MapPost("/start", StartNewRound.HandleAsync)
            .WithName("StartNewRound")
            .Produces<RoundResponse>(StatusCodes.Status201Created)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return gamesGroup;

    }

    private static RouteGroupBuilder MapAbilityEndpoints(this IEndpointRouteBuilder builder)
    {
        var abilitiesGroup = builder.MapGroup("/abilities")
            .WithTags("Ability");

        abilitiesGroup.MapGet("/{id:int}", GetAbility.HandleAsync)
            .WithName("GetAbility")
            .Produces<AbilityResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        abilitiesGroup.MapPatch("/{id:int}", UpdateAbility.HandleAsync)
            .WithName("UpdateAbility")
            .Produces<AbilityResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return abilitiesGroup;
    }

    private static RouteGroupBuilder MapPlayerEndpoints(this IEndpointRouteBuilder builder)
    {
        var playersGroup = builder.MapGroup("/players")
            .WithTags("Player");

        playersGroup.MapGet("/{id:int}", GetPlayer.HandleAsync)
            .WithName("GetPlayerById")
            .Produces<PlayerNameResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapPatch("/{id:int}", UpdatePlayer.HandleAsync)
           .WithName("UpdatePlayer")
           .Produces<PlayerNameResponse>(StatusCodes.Status200OK)
           .ProducesProblem(StatusCodes.Status404NotFound);

        return playersGroup;
    }

    private static RouteGroupBuilder MapRoleEndpoints(this IEndpointRouteBuilder builder)
    {
        var rolesGroup = builder.MapGroup("/roles")
            .WithTags("Role");

        rolesGroup.MapGet("/{id:int}", GetRole.HandleAsync)
            .WithName("GetRole")
            .Produces<RoleResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        rolesGroup.MapPatch("/{id:int}", UpdateRole.HandleAsync)
            .WithName("UpdateRole")
            .Produces<RoleResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return rolesGroup;
    }
}

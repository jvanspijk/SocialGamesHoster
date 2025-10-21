using API.Features.Abilities.Endpoints;
using API.Features.Abilities.Responses;
using API.Features.Authentication.Endpoints;
using API.Features.GameSessions.Endpoints;
using API.Features.Players.Endpoints;
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

        builder.MapPost("/admin/login", AdminLogin.HandleAsync)
            .WithTags("Authentication")
            .WithName("AdminLogin")
            .Produces(StatusCodes.Status200OK)
            .Produces(StatusCodes.Status401Unauthorized);

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
        builder.MapGet("/games/active", GetActiveGameSessions.HandleAsync)
            .WithTags("GameSession")
            .WithName("GetActiveGameId")
            .Produces<int>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);      

        var gamesGroup = builder.MapGroup("/games/{gameId:int}")
           .WithTags("GameSession");

        gamesGroup.MapPatch("/players", UpdateGameParticipants.HandleAsync)
            .WithName("UpdateGameParticipants")
            .Produces<UpdateGameParticipants.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPatch("/ruleset", UpdateGameRuleset.HandleAsync)
            .WithName("UpdateGameRuleset")
            .Produces<UpdateGameRuleset.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/login/{name:alpha}", PlayerLogin.HandleAsync)
            .WithTags("Authentication")
            .WithName("PlayerLogin")
            .Produces<string>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/start", StartGameSession.HandleAsync)
            .WithTags("GameSession")
            .WithName("StartNewGameSession")
            .Produces<StartGameSession.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/end", EndGameSession.HandleAsync)
            .WithName("EndGameSession")
            .Produces(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var playersGroup = gamesGroup.MapGroup("/players")
            .WithTags("Player");

        playersGroup.MapGet("/", GetPlayersFromGame.HandleAsync)
            .WithName("GetGamePlayers")
            .Produces<List<GetPlayersFromGame.Response>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapGet("/{name:alpha}", GetPlayerFromGame.HandleAsync)
            .WithName("GetPlayerByName")
            .Produces<GetPlayerFromGame.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapPost("/", CreatePlayer.HandleAsync)
            .WithName("CreatePlayer")
            .Produces<CreatePlayer.Response>(StatusCodes.Status201Created)
            .Produces<HttpValidationProblemDetails>(StatusCodes.Status400BadRequest);

        var roundsGroup = gamesGroup.MapGroup("rounds")
            .WithTags("Round");

        var currentRoundGroup = roundsGroup.MapGroup("/current")
            .WithTags("Round");

        currentRoundGroup.MapGet("/", GetCurrentRound.HandleAsync)
            .WithName("GetCurrentRound");

        currentRoundGroup.MapPost("/pause", PauseCurrentRound.Handle)
            .WithName("PauseCurrentRound")
            .Produces<TimeSpan>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

        currentRoundGroup.MapPost("/resume", ResumeCurrentRound.Handle)
            .WithName("ResumeCurrentRound")
            .Produces<TimeSpan>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

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

        abilitiesGroup.MapPatch("/{id:int}", UpdateAbilityInformation.HandleAsync)
            .WithName("UpdateAbilityInformation")
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
            .Produces<GetPlayer.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapPatch("/{id:int}", UpdatePlayer.HandleAsync)
           .WithName("UpdatePlayer")
           .Produces<UpdatePlayer.Response>(StatusCodes.Status200OK)
           .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapDelete("/{id:int}", DeletePlayer.HandleAsync)
            .WithName("DeletePlayer")
            .Produces(StatusCodes.Status204NoContent)
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

        rolesGroup.MapPatch("/{id:int}", UpdateRoleInformation.HandleAsync)
            .WithName("UpdateRoleInformation")
            .Produces<RoleResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        rolesGroup.MapPatch("/{id:int}/abilities", UpdateRoleAbilities.HandleAsync)
            .WithName("UpdateRoleAbilities")
            .Produces<RoleResponse>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return rolesGroup;
    }
}

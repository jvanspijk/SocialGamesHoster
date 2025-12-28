using API.Features.Abilities.Endpoints;
using API.Features.Authentication.Endpoints;
using API.Features.Authentication.Hubs;
using API.Features.GameSessions.Endpoints;
using API.Features.Players.Endpoints;
using API.Features.Players.Hubs;
using API.Features.Roles.Endpoints;
using API.Features.Rounds.Endpoints;
using API.Features.Rulesets.Endpoints;

namespace API;

public static class Endpoints
{
    public static IEndpointRouteBuilder MapEndpoints(this IEndpointRouteBuilder builder)
    {        
        builder.MapRoleEndpoints().WithTags("Roles");
        builder.MapAbilityEndpoints().WithTags("Abilities");
        builder.MapPlayerEndpoints()
            .MapHub<PlayersHub>("/hub")
            .WithTags("Players");

        builder.MapGameEndpoints().WithTags("GameSessions");
        builder.MapRulesetEndpoints().WithTags("Rulesets");

        builder.MapPost("/admin/login", AdminLogin.HandleAsync)
            .WithTags("Authentication")
            .WithName(nameof(AdminLogin))
            .Produces<AdminLogin.Response>(StatusCodes.Status200OK)
            .Produces(StatusCodes.Status401Unauthorized);

        builder.MapHub<AuthenticationHub>("/auth/hub");

        return builder;
    }

    private static RouteGroupBuilder MapRulesetEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapGet("/", GetAllRulesets.HandleAsync)
            .WithName(nameof(GetAllRulesets))
            .Produces<List<GetAllRulesets.Response>>(StatusCodes.Status200OK);

        var rulesetsGroup = builder.MapGroup("/rulesets/{rulesetId:int}");

        rulesetsGroup.MapGet("/", GetRuleset.HandleAsync)
            .WithName(nameof(GetRuleset))
            .Produces<GetRuleset.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        rulesetsGroup.MapGet("/full", GetFullRuleset.HandleAsync)
            .WithName(nameof(GetFullRuleset))
            .Produces<GetFullRuleset.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var abilitiesGroup = rulesetsGroup.MapGroup("abilties")
            .WithTags("Abilities");

        abilitiesGroup.MapPost("/", CreateAbility.HandleAsync)
            .WithName(nameof(CreateAbility))  
            .Produces<CreateAbility.Response>(StatusCodes.Status201Created);

        abilitiesGroup.MapGet("/", GetAbilitiesFromRuleset.HandleAsync)
            .WithName(nameof(GetAbilitiesFromRuleset))
            .Produces<List<GetAbilitiesFromRuleset.Response>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var rolesGroup = rulesetsGroup.MapGroup("/roles")
            .WithTags("Roles");

        rolesGroup.MapPost("/", CreateRole.HandleAsync)
            .WithName(nameof(CreateRole))
            .Produces<CreateRole.Response>(StatusCodes.Status201Created);

        rolesGroup.MapGet("/", GetRoles.HandleAsync)
            .WithName(nameof(GetRoles))
            .Produces<List<GetRoles.Response>>(StatusCodes.Status200OK);

        return rulesetsGroup;
    }

    private static RouteGroupBuilder MapGameEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapGet("/games", GetAllGameSessions.HandleAsync)
            .WithName(nameof(GetAllGameSessions))
            .Produces<List<GetAllGameSessions.Response>>(StatusCodes.Status200OK);

        builder.MapGet("/games/active", GetActiveGameSessions.HandleAsync)
            .WithName(nameof(GetActiveGameSessions))
            .Produces<List<GetActiveGameSessions.Response>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);
        
        builder.MapPost("/games/duplicate", DuplicateGameSession.HandleAsync)
            .WithName("DuplicateGameSession")
            .Produces<DuplicateGameSession.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        builder.MapPost("/games", CreateGameSession.HandleAsync)
            .WithName(nameof(CreateGameSession))
            .Produces<CreateGameSession.Response>(StatusCodes.Status200OK);

        var gamesGroup = builder.MapGroup("/games/{gameId:int}");

        gamesGroup.MapGet("/", GetGameSession.HandleAsync)
            .WithName(nameof(GetGameSession))
            .Produces<GetGameSession.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);        

        gamesGroup.MapPatch("/players", UpdateGameParticipants.HandleAsync)
            .WithName(nameof(UpdateGameParticipants))
            .Produces<UpdateGameParticipants.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPatch("/ruleset", UpdateGameRuleset.HandleAsync)
            .WithName(nameof(UpdateGameRuleset))
            .Produces<UpdateGameRuleset.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/winners/add", AddWinners.HandleAsync)
            .WithName(nameof(AddWinners))
            .Produces<AddWinners.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        // wtf is this endpoint doing here?
        gamesGroup.MapPost("/login", PlayerLogin.HandleAsync)
            .WithTags("Authentication")
            .WithName(nameof(PlayerLogin))
            .Produces<PlayerLogin.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/start", StartGameSession.HandleAsync)
            .WithTags("GameSession")
            .WithName(nameof(StartGameSession))
            .Produces<StartGameSession.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/finish", FinishGameSession.HandleAsync)
            .WithName(nameof(FinishGameSession))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/cancel", CancelGameSession.HandleAsync)
            .WithName(nameof(CancelGameSession))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var playersGroup = gamesGroup.MapGroup("/players")
            .WithTags("Player");

        playersGroup.MapGet("/", GetPlayersFromGame.HandleAsync)
            .WithName(nameof(GetPlayersFromGame))
            .Produces<List<GetPlayersFromGame.Response>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapGet("/{playerId:int}", GetPlayerFromGame.HandleAsync)
            .WithName(nameof(GetPlayerFromGame))
            .Produces<GetPlayerFromGame.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapPost("/", CreatePlayer.HandleAsync)
            .WithName(nameof(CreatePlayer))
            .Produces<CreatePlayer.Response>(StatusCodes.Status201Created)
            .Produces<HttpValidationProblemDetails>(StatusCodes.Status400BadRequest);

        var roundsGroup = gamesGroup.MapGroup("rounds")
            .WithTags("Rounds");

        var currentRoundGroup = roundsGroup.MapGroup("/current")
            .WithTags("Rounds");

        currentRoundGroup.MapGet("/", GetCurrentRound.HandleAsync)
            .WithName(nameof(GetCurrentRound))
            .Produces<GetCurrentRound.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        currentRoundGroup.MapPatch("/time", AdjustRoundTime.HandleAsync)
            .WithName(nameof(AdjustRoundTime))
            .WithDescription("Adds or removes time from the round timer.")
            .Produces<TimeSpan>(StatusCodes.Status200OK)            
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        currentRoundGroup.MapPost("/pause", PauseCurrentRound.Handle)
            .WithName(nameof(PauseCurrentRound))
            .Produces<TimeSpan>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

        currentRoundGroup.MapPost("/resume", ResumeCurrentRound.Handle)
            .WithName(nameof(ResumeCurrentRound))
            .Produces<TimeSpan>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

        currentRoundGroup.MapPost("/cancel", CancelCurrentRound.Handle)
            .WithName(nameof(CancelCurrentRound))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        currentRoundGroup.MapPost("/finish", FinishCurrentRound.Handle)
            .WithName(nameof(FinishCurrentRound))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        roundsGroup.MapPost("/", StartNewRound.HandleAsync)
            .WithName(nameof(StartNewRound))
            .Produces<StartNewRound.Response>(StatusCodes.Status201Created)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return gamesGroup;
    }

    private static RouteGroupBuilder MapAbilityEndpoints(this IEndpointRouteBuilder builder)
    {
        var abilitiesGroup = builder.MapGroup("/abilities");

        abilitiesGroup.MapGet("/{id:int}", GetAbility.HandleAsync)
            .WithName(nameof(GetAbility))
            .Produces<GetAbility.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        abilitiesGroup.MapPatch("/{id:int}", UpdateAbilityInformation.HandleAsync)
            .WithName(nameof(UpdateAbilityInformation))
            .Produces<UpdateAbilityInformation.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return abilitiesGroup;
    }

    private static RouteGroupBuilder MapPlayerEndpoints(this IEndpointRouteBuilder builder)
    {
        var playersGroup = builder.MapGroup("/players")
            .WithTags("Player");

        playersGroup.MapGet("/{id:int}", GetPlayer.HandleAsync)
            .WithName(nameof(GetPlayer))
            .Produces<GetPlayer.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapGet("/{id:int}/full", GetFullPlayer.HandleAsync)
            .WithName(nameof(GetFullPlayer))
            .Produces<GetFullPlayer.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapGet("/{id:int}/{gameId:int}/exists", PlayerExistsInGame.HandleAsync)
            .WithName(nameof(PlayerExistsInGame))
            .Produces<PlayerExistsInGame.Response>(StatusCodes.Status200OK);

        playersGroup.MapPatch("/{id:int}", UpdatePlayer.HandleAsync)
           .WithName(nameof(UpdatePlayer))
           .Produces<UpdatePlayer.Response>(StatusCodes.Status200OK)
           .ProducesProblem(StatusCodes.Status404NotFound);

        playersGroup.MapDelete("/{id:int}", DeletePlayer.HandleAsync)
            .WithName(nameof(DeletePlayer))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return playersGroup;
    }

    private static RouteGroupBuilder MapRoleEndpoints(this IEndpointRouteBuilder builder)
    {
        var rolesGroup = builder.MapGroup("/roles");

        rolesGroup.MapGet("/{id:int}", GetRole.HandleAsync)
            .WithName(nameof(GetRole))
            .Produces<GetRole.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        rolesGroup.MapPatch("/{id:int}", UpdateRoleInformation.HandleAsync)
            .WithName(nameof(UpdateRoleInformation))
            .Produces<UpdateRoleInformation.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        rolesGroup.MapPatch("/{id:int}/abilities", UpdateRoleAbilities.HandleAsync)
            .WithName(nameof(UpdateRoleAbilities))
            .Produces<UpdateRoleAbilities.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return rolesGroup;
    }
}

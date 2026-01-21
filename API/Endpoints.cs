using API.Features.Abilities.Endpoints;
using API.Features.Abilities.Hubs;
using API.Features.Auth.Endpoints;
using API.Features.Auth.Hubs;
using API.Features.Chat.Endpoints;
using API.Features.Chat.Hubs;
using API.Features.GameSessions.Endpoints;
using API.Features.GameSessions.Hubs;
using API.Features.Players.Endpoints;
using API.Features.Players.Hubs;
using API.Features.Roles.Endpoints;
using API.Features.Roles.Hubs;
using API.Features.Rounds.Endpoints;
using API.Features.Rounds.Hubs;
using API.Features.Rulesets.Endpoints;
using API.Features.Rulesets.Hubs;
using API.Features.Timers.Endpoints;
using API.Features.Timers.Hubs;

namespace API;

public static class Endpoints
{
    public static IEndpointRouteBuilder MapEndpoints(this IEndpointRouteBuilder builder)
    {
        builder.MapRoleEndpoints()
            .MapHub<RolesHub>("/hub")
            .WithTags("Roles");

        builder.MapAbilityEndpoints()
            .MapHub<AbilitiesHub>("/hub")
            .WithTags("Abilities");

        builder.MapPlayerEndpoints()
            .MapHub<PlayersHub>("/hub")
            .WithTags("Players");

        builder.MapGameEndpoints()
            .MapHub<GameSessionsHub>("/hub")
            .WithTags("GameSessions");

        builder.MapRulesetEndpoints()
            .MapHub<RulesetsHub>("/hub")
            .WithTags("Rulesets");

        builder.MapTimerEndpoints()
            .MapHub<TimersHub>("/hub")
            .WithTags("Timers");

        builder.MapRoundEndpoints()
            .MapHub<RoundsHub>("/hub")
            .WithTags("Rounds");

        builder.MapAuthEndpoints()
            .MapHub<AuthenticationHub>("/hub")
            .WithTags("Auth");

        builder.MapChatEndpoints()
            .MapHub<ChatHub>("/hub")
            .WithTags("Chat");

        return builder;
    }

    private static RouteGroupBuilder MapRoundEndpoints(this IEndpointRouteBuilder builder)
    {
        var roundsGroup = builder.MapGroup("/rounds");
        //roundsGroup.MapGet("/{id:int}", GetRound.HandleAsync)
        //    .WithName(nameof(GetRound))
        //    .Produces<GetRound.Response>(StatusCodes.Status200OK)
        //    .ProducesProblem(StatusCodes.Status404NotFound);
        var currentRoundGroup = roundsGroup.MapGroup("/current")
          .WithTags("Rounds");

        currentRoundGroup.MapGet("/", GetCurrentRound.HandleAsync)
            .WithName(nameof(GetCurrentRound))
            .Produces<GetCurrentRound.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        currentRoundGroup.MapPost("/finish", FinishCurrentRound.HandleAsync)
            .WithName(nameof(FinishCurrentRound))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        roundsGroup.MapPost("/", StartNewRound.HandleAsync)
            .WithName(nameof(StartNewRound))
            .Produces<StartNewRound.Response>(StatusCodes.Status201Created)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return roundsGroup;
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

        gamesGroup.MapPost("/start", StartGameSession.HandleAsync)
            .WithTags("GameSession")
            .WithName(nameof(StartGameSession))
            .Produces<StartGameSession.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/stop", StopGameSession.HandleAsync)
            .WithName(nameof(StopGameSession))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status404NotFound);

        gamesGroup.MapPost("/delete", DeleteGameSession.HandleAsync)
            .WithName(nameof(DeleteGameSession))
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

        return gamesGroup;
    }

    private static RouteGroupBuilder MapTimerEndpoints(this IEndpointRouteBuilder builder)
    {
        var timersGroup = builder.MapGroup("/timers");

        timersGroup.MapGet("/", GetTimerState.Handle)
            .WithName(nameof(GetTimerState))
            .Produces<GetTimerState.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound)
            .DisableClientCaching();

        timersGroup.MapPut("/adjust", AdjustTimer.HandleAsync)
           .WithName(nameof(AdjustTimer))
           .Produces<AdjustTimer.Response>(StatusCodes.Status200OK)
           .ProducesProblem(StatusCodes.Status400BadRequest);

        timersGroup.MapPost("/start", StartTimer.HandleAsync)
            .WithName(nameof(StartTimer))
            .Produces<StartTimer.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

        timersGroup.MapPost("/pause", PauseTimer.HandleAsync)
            .WithName(nameof(PauseTimer))
            .Produces<PauseTimer.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

        timersGroup.MapPost("/resume", ResumeTimer.HandleAsync)
            .WithName(nameof(ResumeTimer))
            .Produces<ResumeTimer.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest);

        timersGroup.MapPost("/stop", StopTimer.HandleAsync)
            .WithName(nameof(StopTimer))
            .Produces(StatusCodes.Status204NoContent)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        return timersGroup;
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


    public static RouteGroupBuilder MapAuthEndpoints(this IEndpointRouteBuilder builder)
    {
        var authGroup = builder.MapGroup("/auth");

        authGroup.MapPost("/admin/login", AdminLogin.HandleAsync)
            .WithName(nameof(AdminLogin))
            .Produces<AdminLogin.Response>(StatusCodes.Status200OK)
            .Produces(StatusCodes.Status401Unauthorized);

        authGroup.MapPost("/player/login", PlayerLogin.HandleAsync)
            .WithName(nameof(PlayerLogin))
            .Produces<PlayerLogin.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        authGroup.MapGet("/me", Me.HandleAsync)
            .WithName(nameof(Me))
            .Produces<Me.Response>(StatusCodes.Status200OK);

        return authGroup;
    }

    public static RouteGroupBuilder MapChatEndpoints(this IEndpointRouteBuilder builder)
    {
        var chatGroup = builder.MapGroup("/chat");

        chatGroup.MapGet("/{id:guid}", GetMessage.HandleAsync)
            .WithName(nameof(GetMessage))
            .Produces<GetMessage.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound);

        var channelGroup = chatGroup.MapGroup("/channels/{channelId:int}");

        channelGroup.MapPost("/", CreateChannel.HandleAsync)
            .WithName(nameof(CreateChannel))
            .Produces<CreateChannel.Response>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        //channelGroup.MapPost("/join", JoinChannel.HandleAsync)
        //    .WithName(nameof(JoinChannel))
        //    .Produces(StatusCodes.Status204NoContent)
        //    .ProducesProblem(StatusCodes.Status400BadRequest)
        //    .ProducesProblem(StatusCodes.Status404NotFound);

        channelGroup.MapPost("/send", SendMessage.HandleAsync)
            .WithName(nameof(SendMessage))
            .Produces(StatusCodes.Status201Created)
            .ProducesProblem(StatusCodes.Status400BadRequest)
            .ProducesProblem(StatusCodes.Status404NotFound);

        channelGroup.MapGet("/messages", GetMessagesFromChannel.HandleAsync)
            .WithName(nameof(GetMessagesFromChannel))
            .Produces<List<GetMessagesFromChannel.Response>>(StatusCodes.Status200OK)
            .ProducesProblem(StatusCodes.Status404NotFound)
            .DisableClientCaching();

        return chatGroup;
    }
}



public static class RouteHandlerBuilderExtensions
{
    public static RouteHandlerBuilder DisableClientCaching(this RouteHandlerBuilder builder)
    {
        builder.AddEndpointFilter(async (context, next) =>
        {
            var headers = context.HttpContext.Response.Headers;
            headers.CacheControl = "no-store, no-cache, must-revalidate";
            headers.Pragma = "no-cache";
            headers.Expires = "0";
            return await next(context);
        });
        return builder;
    }

    public static RouteHandlerBuilder WithClientCaching(this RouteHandlerBuilder builder, int seconds)
    {
        return builder.AddEndpointFilter(async (context, next) =>
        {
            var headers = context.HttpContext.Response.Headers;
            headers.CacheControl = $"public, max-age={seconds}";
            return await next(context);
        });
    }
}

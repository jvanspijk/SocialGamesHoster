package games

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
)

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1")

	group.GET("/games", listGames).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games", createGame).BindFunc(actorauth.RequireGameMaster)
	group.GET("/games/{id}/admin-view", adminView).BindFunc(actorauth.RequireGameMaster)
	group.GET("/games/live/player-view", playerView).BindFunc(actorauth.RequirePlayer)
	group.POST("/games/{id}/duplicate", duplicateGame).BindFunc(actorauth.RequireGameMaster)
	group.DELETE("/games/{id}", deleteGame).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/open-lobby", openLobby).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/cancel", cancelGame).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/join", joinGame).BindFunc(actorauth.RequirePlayer)
	group.POST("/games/{id}/start", transition(Start)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/pause", transition(Pause)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/resume", transition(Resume)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/review", transition(BeginReview)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/return-to-running", transition(ReturnToRunning)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/archive", archiveGame).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/phase", setPhase).BindFunc(actorauth.RequireGameMaster)
	group.PATCH("/games/{id}/role-visibility", setRoleVisibility).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/completion/start", startCompletion).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/completion/cancel", cancelCompletion).BindFunc(actorauth.RequireGameMaster)
	group.GET("/games/{id}/summary", gameSummary).BindFunc(actorauth.RequireGameMaster)
	group.GET("/games/{id}/activity", listActivity).BindFunc(actorauth.RequireGameMaster)

	group.PATCH("/games/{id}/participants/{participantId}", updateParticipant).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/participants/{participantId}/kick", setParticipantStatus(gamepolicy.ParticipantKicked)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/participants/{participantId}/eliminate", setParticipantStatus(gamepolicy.ParticipantEliminated)).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/participants/{participantId}/reinstate", setParticipantStatus(gamepolicy.ParticipantActive)).BindFunc(actorauth.RequireGameMaster)
	group.PUT("/games/{id}/assignments", putAssignments).BindFunc(actorauth.RequireGameMaster)
	group.POST("/games/{id}/assignments/randomize", randomizeAssignments).BindFunc(actorauth.RequireGameMaster)
	group.PUT("/games/{id}/outcomes", putOutcomes).BindFunc(actorauth.RequireGameMaster)

	group.GET("/games/live", func(event *core.RequestEvent) error {
		game, err := findLiveGame(event.App)
		if err != nil {
			return event.JSON(http.StatusNotFound, map[string]string{"code": "game.no_live_game", "message": "There is no live game."})
		}
		return event.JSON(http.StatusOK, projectGame(game))
	})
}

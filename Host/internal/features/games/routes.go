package games

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	platformauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/auth"
)

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1")

	group.GET("/games", listGames).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games", createGame).BindFunc(platformauth.RequireGameMaster)
	group.GET("/games/{id}/admin-view", adminView).BindFunc(platformauth.RequireGameMaster)
	group.GET("/games/live/player-view", playerView).BindFunc(platformauth.RequirePlayer)
	group.POST("/games/{id}/duplicate", duplicateGame).BindFunc(platformauth.RequireGameMaster)
	group.DELETE("/games/{id}", deleteGame).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/open-lobby", openLobby).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/open-joining", openJoining).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/close-joining", closeJoining).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/join", joinGame).BindFunc(platformauth.RequirePlayer)
	group.POST("/games/{id}/start", transition(Start)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/pause", transition(Pause)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/resume", transition(Resume)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/review", transition(BeginReview)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/return-to-running", transition(ReturnToRunning)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/archive", archiveGame).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/phase", setPhase).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/announcements", createAnnouncement).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/announcements/{announcementId}/acknowledge", acknowledgeAnnouncement).BindFunc(platformauth.RequirePlayer)
	group.GET("/games/{id}/announcements/{announcementId}/media/{kind}", announcementMedia)
	group.PATCH("/games/{id}/role-visibility", setRoleVisibility).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/completion/start", startCompletion).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/completion/cancel", cancelCompletion).BindFunc(platformauth.RequireGameMaster)
	group.GET("/games/{id}/summary", gameSummary).BindFunc(platformauth.RequireGameMaster)
	group.GET("/games/{id}/activity", listActivity).BindFunc(platformauth.RequireGameMaster)
	group.GET("/games/{id}/announcements", listAnnouncements).BindFunc(platformauth.RequireGameMaster)

	group.PATCH("/games/{id}/participants/{participantId}", updateParticipant).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/participants/{participantId}/kick", setParticipantStatus(gamepolicy.ParticipantKicked)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/participants/{participantId}/eliminate", setParticipantStatus(gamepolicy.ParticipantEliminated)).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/participants/{participantId}/reinstate", setParticipantStatus(gamepolicy.ParticipantActive)).BindFunc(platformauth.RequireGameMaster)
	group.PUT("/games/{id}/assignments", putAssignments).BindFunc(platformauth.RequireGameMaster)
	group.POST("/games/{id}/assignments/randomize", randomizeAssignments).BindFunc(platformauth.RequireGameMaster)
	group.PUT("/games/{id}/outcomes", putOutcomes).BindFunc(platformauth.RequireGameMaster)

	group.GET("/games/live", func(event *core.RequestEvent) error {
		game, err := findLiveGame(event.App)
		if err != nil {
			return event.JSON(http.StatusNotFound, map[string]string{"code": "game.no_live_game", "message": "There is no live game."})
		}
		return event.JSON(http.StatusOK, projectGame(game))
	})
}

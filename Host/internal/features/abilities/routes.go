package abilities

import (
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func Register(event *core.ServeEvent) {
	group := event.Router.Group("/api/app/v1")
	group.POST("/games/{id}/abilities/{abilityId}/activate", activate).BindFunc(actorauth.RequirePlayer)
	group.DELETE("/games/{id}/abilities/{abilityId}/activate", undo).BindFunc(actorauth.RequirePlayer)
}

func activate(event *core.RequestEvent) error {
	choices, err := Activate(event.App, event.Request.PathValue("id"), event.Auth.Id, event.Request.PathValue("abilityId"), time.Now().UTC())
	if err != nil {
		return writeError(event, err)
	}
	publishChanged(event.App, event.Request.PathValue("id"))
	return event.JSON(http.StatusOK, map[string]any{"choices": choices})
}

func undo(event *core.RequestEvent) error {
	choices, err := Undo(event.App, event.Request.PathValue("id"), event.Auth.Id, event.Request.PathValue("abilityId"))
	if err != nil {
		return writeError(event, err)
	}
	publishChanged(event.App, event.Request.PathValue("id"))
	return event.JSON(http.StatusOK, map[string]any{"choices": choices})
}

func publishChanged(app core.App, gameID string) {
	game, err := app.FindRecordById("games", gameID)
	if err != nil {
		return
	}
	_ = realtime.Publish(app, "game:"+gameID+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: gameID, Revision: game.GetInt("revision"),
		Kind: "ability.progress_changed", Payload: map[string]any{},
	}, func(auth *core.Record) bool {
		return auth != nil && auth.GetBool("active")
	})
}

func writeError(event *core.RequestEvent, err error) error {
	if appError, ok := err.(result.AppError); ok {
		return httpx.WriteError(event, appError)
	}
	return httpx.WriteError(event, result.Internal(err))
}

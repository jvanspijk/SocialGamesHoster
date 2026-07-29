package audit

import (
	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
)

func Record(
	app core.App,
	actor *core.Record,
	gameID string,
	action string,
	targetType string,
	targetID string,
	detail any,
	traceID any,
) error {
	collection, err := app.FindCollectionByNameOrId("game_audit")
	if err != nil {
		return err
	}
	record := core.NewRecord(collection)
	record.Set("game", gameID)
	record.Set("actor_type", "system")
	record.Set("actor_label", "System")
	if actor != nil {
		record.Set("actor_id", actor.Id)
		record.Set("actor_label", actor.GetString("display_name"))
		if actorauth.IsGameMaster(actor) {
			record.Set("actor_type", "game_master")
		} else {
			record.Set("actor_type", "player")
		}
	}
	record.Set("action", action)
	record.Set("target_type", targetType)
	record.Set("target_id", targetID)
	record.Set("detail", detail)
	if trace, ok := traceID.(string); ok {
		record.Set("request_id", trace)
	}
	return app.Save(record)
}

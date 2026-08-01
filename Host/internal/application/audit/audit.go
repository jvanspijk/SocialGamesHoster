// Package audit writes immutable application audit entries.
package audit

import (
	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
)

// Record persists one audit entry using the supplied app, which may be the
// base app or a transaction app. Feature slices choose the action, target, and
// safe detail payload; this package snapshots the actor at write time.
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
	if actorauth.IsGameMaster(actor) {
		record.Set("actor_id", actor.Id)
		record.Set("actor_type", "game_master")
		record.Set("actor_label", actor.GetString("display_name"))
	} else if actorauth.IsPlayer(actor) {
		record.Set("actor_id", actor.Id)
		record.Set("actor_type", "player")
		record.Set("actor_label", actor.GetString("display_name"))
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

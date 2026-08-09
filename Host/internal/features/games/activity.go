package games

import (
	"net/http"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/application/pagination"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func listActivity(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	filter := "game = {:game}"
	params := dbx.Params{"game": game.Id}
	if cursor := event.Request.URL.Query().Get("cursor"); cursor != "" {
		position, err := pagination.Decode(cursor)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("activity.invalid_cursor", "The activity cursor is invalid.", nil))
		}
		filter += " && " + pagination.DescendingCreatedIDPredicate
		params["created"] = position.Created
		params["id"] = position.ID
	}
	records, err := event.App.FindRecordsByFilter("game_audit", filter, pagination.DescendingCreatedIDSort, pagination.QueryLimit, 0, params)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	records, nextCursor := pagination.Window(records, func(record *core.Record) pagination.Cursor {
		return pagination.Cursor{Created: record.GetDateTime("created").Time().UTC(), ID: record.Id}
	})
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		label := activityLabel(record.GetString("actor_label"), record.GetString("action"))
		if label == "" {
			continue
		}
		items = append(items, map[string]any{
			"id": record.Id, "text": label, "createdAt": dateValue(record, "created"),
		})
	}
	return event.JSON(http.StatusOK, map[string]any{"items": items, "nextCursor": nextCursor})
}

func activityLabel(actor, action string) string {
	if strings.TrimSpace(actor) == "" {
		actor = "A game master"
	}
	phrases := map[string]string{
		"game.pause":                    "paused the game",
		"game.resume":                   "resumed the game",
		"game.start":                    "started the game",
		"game.open_lobby":               "opened the lobby",
		"game.joining_opened":           "reopened joining",
		"game.joining_closed":           "closed joining",
		"game.roles_available":          "made roles available",
		"game.roles_hidden":             "hid player roles",
		"game.phase_changed":            "changed the phase",
		"game.completion_started":       "started the completion flow",
		"game.completion_cancelled":     "returned to the game",
		"chat.players_can_post_changed": "changed who can post in chat",
		"outcomes.changed":              "updated player outcomes",
		"assignments.changed":           "updated role assignments",
		"game.archived":                 "finished and archived the game",
	}
	phrase := phrases[action]
	if phrase == "" {
		return ""
	}
	return actor + " " + phrase
}

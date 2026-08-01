package games

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

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
		created, id, err := decodeActivityCursor(cursor)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("activity.invalid_cursor", "The activity cursor is invalid.", nil))
		}
		filter += " && (created < {:created} || (created = {:created} && id < {:id}))"
		params["created"] = created
		params["id"] = id
	}
	records, err := event.App.FindRecordsByFilter("game_audit", filter, "-created,-id", 51, 0, params)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	hasMore := len(records) > 50
	if hasMore {
		records = records[:50]
	}
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
	nextCursor := ""
	if hasMore && len(records) > 0 {
		last := records[len(records)-1]
		nextCursor = encodeActivityCursor(last.GetDateTime("created").Time().UTC(), last.Id)
	}
	return event.JSON(http.StatusOK, map[string]any{"items": items, "nextCursor": nextCursor})
}

type activityCursor struct {
	Created time.Time `json:"created"`
	ID      string    `json:"id"`
}

func encodeActivityCursor(created time.Time, id string) string {
	data, _ := json.Marshal(activityCursor{Created: created, ID: id})
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeActivityCursor(value string) (time.Time, string, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, "", err
	}
	var cursor activityCursor
	if err := json.Unmarshal(data, &cursor); err != nil || cursor.Created.IsZero() || cursor.ID == "" {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	return cursor.Created, cursor.ID, nil
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

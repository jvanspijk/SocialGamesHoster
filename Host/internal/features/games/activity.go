package games

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func listActivity(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	offset, err := decodeOffset(event.Request.URL.Query().Get("cursor"))
	if err != nil {
		return httpx.WriteError(event, result.Invalid("activity.invalid_cursor", "The activity cursor is invalid.", nil))
	}
	records, err := event.App.FindRecordsByFilter("game_audit", "game = {:game}", "-created,-id", 51, offset, dbx.Params{"game": game.Id})
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
	if hasMore {
		nextCursor = encodeOffset(offset + 50)
	}
	return event.JSON(http.StatusOK, map[string]any{"items": items, "nextCursor": nextCursor})
}

func listAnnouncements(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return writeGameError(event, err)
	}
	offset, err := decodeOffset(event.Request.URL.Query().Get("cursor"))
	if err != nil {
		return httpx.WriteError(event, result.Invalid("announcement.invalid_cursor", "The announcement cursor is invalid.", nil))
	}
	records, err := event.App.FindRecordsByFilter("attention_items", "game = {:game}", "-created,-id", 51, offset, dbx.Params{"game": game.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	hasMore := len(records) > 50
	if hasMore {
		records = records[:50]
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if summary, err := projectAdminAttentionSummary(event.App, record); err == nil {
			items = append(items, summary)
		}
	}
	nextCursor := ""
	if hasMore {
		nextCursor = encodeOffset(offset + 50)
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

func encodeOffset(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeOffset(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return 0, err
	}
	offset, err := strconv.Atoi(string(decoded))
	if err != nil || offset < 0 {
		return 0, strconv.ErrSyntax
	}
	return offset, nil
}

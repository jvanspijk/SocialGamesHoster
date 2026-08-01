package chat

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func listAnnouncements(event *core.RequestEvent) error {
	game, err := findAnnouncementGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	filter := "game = {:game}"
	params := dbx.Params{"game": game.Id}
	if cursor := event.Request.URL.Query().Get("cursor"); cursor != "" {
		created, id, err := decodeAnnouncementCursor(cursor)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("announcement.invalid_cursor", "The announcement cursor is invalid.", nil))
		}
		filter += " && (created < {:created} || (created = {:created} && id < {:id}))"
		params["created"] = created
		params["id"] = id
	}
	records, err := event.App.FindRecordsByFilter("attention_items", filter, "-created,-id", 51, 0, params)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	hasMore := len(records) > 50
	if hasMore {
		records = records[:50]
	}
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if summary, summaryErr := projectAdminAttentionSummary(event.App, record); summaryErr == nil {
			items = append(items, summary)
		}
	}
	nextCursor := ""
	if hasMore && len(records) > 0 {
		last := records[len(records)-1]
		nextCursor = encodeAnnouncementCursor(last.GetDateTime("created").Time().UTC(), last.Id)
	}
	return event.JSON(http.StatusOK, map[string]any{"items": items, "nextCursor": nextCursor})
}

type announcementCursor struct {
	Created time.Time `json:"created"`
	ID      string    `json:"id"`
}

func encodeAnnouncementCursor(created time.Time, id string) string {
	data, _ := json.Marshal(announcementCursor{Created: created, ID: id})
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeAnnouncementCursor(value string) (time.Time, string, error) {
	data, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil {
		return time.Time{}, "", err
	}
	var cursor announcementCursor
	if err := json.Unmarshal(data, &cursor); err != nil || cursor.Created.IsZero() || cursor.ID == "" {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	return cursor.Created, cursor.ID, nil
}

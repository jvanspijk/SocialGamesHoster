package chat

import (
	"encoding/base64"
	"net/http"
	"strconv"

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
	offset, err := decodeAnnouncementOffset(event.Request.URL.Query().Get("cursor"))
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
		if summary, summaryErr := projectAdminAttentionSummary(event.App, record); summaryErr == nil {
			items = append(items, summary)
		}
	}
	nextCursor := ""
	if hasMore {
		nextCursor = encodeAnnouncementOffset(offset + 50)
	}
	return event.JSON(http.StatusOK, map[string]any{"items": items, "nextCursor": nextCursor})
}

func encodeAnnouncementOffset(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}

func decodeAnnouncementOffset(value string) (int, error) {
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

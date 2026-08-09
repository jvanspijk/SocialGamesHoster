package chat

import (
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/application/pagination"
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
		position, err := pagination.Decode(cursor)
		if err != nil {
			return httpx.WriteError(event, result.Invalid("announcement.invalid_cursor", "The announcement cursor is invalid.", nil))
		}
		filter += " && " + pagination.DescendingCreatedIDPredicate
		params["created"] = position.Created
		params["id"] = position.ID
	}
	records, err := event.App.FindRecordsByFilter("attention_items", filter, pagination.DescendingCreatedIDSort, pagination.QueryLimit, 0, params)
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	records, nextCursor := pagination.Window(records, func(record *core.Record) pagination.Cursor {
		return pagination.Cursor{Created: record.GetDateTime("created").Time().UTC(), ID: record.Id}
	})
	items := make([]map[string]any, 0, len(records))
	for _, record := range records {
		if summary, summaryErr := projectAdminAttentionSummary(event.App, record); summaryErr == nil {
			items = append(items, summary)
		}
	}
	return event.JSON(http.StatusOK, map[string]any{"items": items, "nextCursor": nextCursor})
}

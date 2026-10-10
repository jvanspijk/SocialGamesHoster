package chat

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

func liveAnnouncement(event *core.RequestEvent) (*core.Record, *core.Record, error) {
	game, err := findAnnouncementGame(event)
	if err != nil {
		return nil, nil, err
	}
	status := gamepolicy.GameStatus(game.GetString("status"))
	if status != gamepolicy.GameLobby && status != gamepolicy.GameRunning && status != gamepolicy.GamePaused {
		return nil, nil, result.Conflict("chat.read_only", "Announcements are only available during a live game.")
	}
	item, err := event.App.FindRecordById("attention_items", event.Request.PathValue("announcementId"))
	if err != nil || item.GetString("game") != game.Id || item.GetString("kind") != "announcement" {
		return nil, nil, result.AppError{Code: "attention.not_found", Message: "Announcement not found.", Status: http.StatusNotFound}
	}
	return game, item, nil
}

func remindAnnouncement(event *core.RequestEvent) error {
	game, item, err := liveAnnouncement(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	var summary map[string]any
	err = event.App.RunInTransaction(func(tx core.App) error {
		current, err := tx.FindRecordById("attention_items", item.Id)
		if err != nil {
			return err
		}
		if current.GetBool("tracking_cleared") {
			return result.Conflict("attention.tracking_cleared", "This announcement has been cleared.")
		}
		summary, err = projectAdminAttentionSummary(tx, current)
		if err != nil {
			return err
		}
		if len(summary["unacknowledgedPlayers"].([]map[string]any)) == 0 {
			return result.Conflict("attention.no_pending_recipients", "No players are waiting to acknowledge this announcement.")
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "attention.announcement_reminded", "attention_item", item.Id,
			map[string]any{"recipientTotal": len(summary["unacknowledgedPlayers"].([]map[string]any))}, event.Get(httpx.TraceIDKey))
	})
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	// Reuse the saved announcement and receipts, including its private media.
	// Recheck each receipt at publication so an acknowledgement cannot be reset.
	allowed := func(auth *core.Record) bool {
		return actorHasAnnouncementReceipt(event.App, item.Id, auth, true)
	}
	_ = realtime.Publish(event.App, "game:"+game.Id+":public", realtime.Event[any]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: "attention.announcement_reminded", Payload: projectPlayerAttention(item),
	}, allowed)
	if item.GetString("cue_key") != "" {
		definition, err := definitionFromGame(game)
		if err == nil {
			for _, cue := range definition.AudioCues {
				if cue.ID != item.GetString("cue_key") {
					continue
				}
				publishAttentionCue(event.App, game, cue, item.Id, true)
			}
		}
	}
	publishAnnouncementGameMasters(event.App, game, "attention.announcement_reminded", summary)
	return event.JSON(http.StatusOK, summary)
}

func clearAnnouncementTracking(event *core.RequestEvent) error {
	game, item, err := liveAnnouncement(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	err = event.App.RunInTransaction(func(tx core.App) error {
		current, err := tx.FindRecordById("attention_items", item.Id)
		if err != nil || current.GetBool("tracking_cleared") {
			return err
		}
		current.Set("tracking_cleared", true)
		if err := tx.Save(current); err != nil {
			return err
		}
		return applicationaudit.Record(tx, event.Auth, game.Id, "attention.announcement_tracking_cleared", "attention_item", item.Id, nil, event.Get(httpx.TraceIDKey))
	})
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	publishAnnouncementGameMasters(event.App, game, "attention.announcement_tracking_cleared", map[string]any{"id": item.Id})
	return event.NoContent(http.StatusNoContent)
}

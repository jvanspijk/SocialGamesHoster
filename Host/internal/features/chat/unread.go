package chat

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/types"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

const unreadCountLimit = 99

type readMarker struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

// Both reads and counts use the same frozen historical membership window.
func messageHistoryWindow(resolved access) (time.Time, time.Time, bool) {
	if resolved.IsGM || resolved.Membership == nil || resolved.Membership.GetDateTime("left_at").IsZero() ||
		(rulesets.RoleControlsChatRoom(resolved.Room.GetString("kind")) && !gamepolicy.IsArchived(gamepolicy.GameStatus(resolved.Game.GetString("status")))) {
		return time.Time{}, time.Time{}, false
	}
	return resolved.Membership.GetDateTime("joined_at").Time().UTC(), resolved.Membership.GetDateTime("left_at").Time().UTC(), true
}

func unreadCounts(event *core.RequestEvent) error {
	if !actorauth.IsActiveGameMaster(event.Auth) && !actorauth.IsActivePlayer(event.Auth) {
		return httpx.WriteError(event, result.AppError{Code: "auth.required", Message: "Sign in to continue.", Status: http.StatusUnauthorized})
	}
	game, err := event.App.FindRecordById("games", event.Request.PathValue("id"))
	if err != nil {
		return httpx.WriteError(event, result.AppError{Code: "game.not_found", Message: "Game not found.", Status: http.StatusNotFound})
	}
	var request struct {
		Markers map[string]readMarker `json:"markers"`
	}
	if err := event.BindBody(&request); err != nil || len(request.Markers) > 200 {
		return httpx.WriteError(event, result.Invalid("chat.invalid_markers", "The chat read positions are invalid.", nil))
	}
	for _, marker := range request.Markers {
		if marker.ID == "" || len(marker.ID) > 128 || marker.CreatedAt.IsZero() {
			return httpx.WriteError(event, result.Invalid("chat.invalid_markers", "The chat read positions are invalid.", nil))
		}
	}
	rooms, err := event.App.FindRecordsByFilter("chat_rooms", "game = {:game} && kind != 'announcements'", "kind,label", 200, 0, dbx.Params{"game": game.Id})
	if err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	counts := map[string]int{}
	queries := []string{}
	params := dbx.Params{}
	accesses, err := resolveGameRoomAccess(event, game, rooms)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	for index, resolved := range accesses {
		if !resolved.IsGM && !resolved.Policy.Readable {
			continue
		}
		room := resolved.Room
		counts[room.Id] = 0
		key := fmt.Sprintf("r%d", index)
		params[key] = room.Id
		predicate := "room = {:" + key + "}"
		if marker, ok := request.Markers[room.Id]; ok {
			params[key+"time"] = marker.CreatedAt.UTC().Format(types.DefaultDateLayout)
			params[key+"id"] = marker.ID
			predicate += " AND (created > {:" + key + "time} OR (created = {:" + key + "time} AND id > {:" + key + "id}))"
		}
		if joined, left, bounded := messageHistoryWindow(resolved); bounded {
			params[key+"joined"] = joined.Format(types.DefaultDateLayout)
			params[key+"left"] = left.Format(types.DefaultDateLayout)
			predicate += " AND created >= {:" + key + "joined} AND created <= {:" + key + "left}"
		}
		// Count from the existing room/created/id index, bounded independently per room.
		queries = append(queries, "SELECT {:"+key+"} AS room, COUNT(*) AS count FROM (SELECT id FROM chat_messages WHERE "+predicate+" LIMIT 99)")
	}
	var rows []struct {
		Room  string `db:"room"`
		Count int    `db:"count"`
	}
	if len(queries) > 0 {
		if err := event.App.DB().NewQuery(strings.Join(queries, " UNION ALL ")).Bind(params).All(&rows); err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
	}
	total := 0
	for _, row := range rows {
		counts[row.Room] = row.Count
		total += row.Count
	}
	if total > unreadCountLimit {
		total = unreadCountLimit
	}
	return event.JSON(http.StatusOK, map[string]any{"counts": counts, "total": total})
}

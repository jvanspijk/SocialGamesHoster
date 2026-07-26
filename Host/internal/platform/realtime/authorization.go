package realtime

import (
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func RegisterAuthorization(app core.App) {
	app.OnRealtimeSubscribeRequest().BindFunc(func(event *core.RealtimeSubscribeRequestEvent) error {
		for _, topic := range event.Subscriptions {
			if !canSubscribe(app, event.Auth, topic) {
				return errors.New("this realtime subscription is not available")
			}
		}
		return event.Next()
	})
}

func canSubscribe(app core.App, auth *core.Record, topic string) bool {
	if requestID, capability, ok := splitProfileRequestTopic(topic); ok {
		return profileRequestCapabilityMatches(app, requestID, capability)
	}
	kind, id, ok := splitTopic(topic)
	if !ok {
		return false
	}
	if activeGameMaster(auth) {
		switch kind {
		case "game", "game-master", "participant", "room", "profile":
			return recordExists(app, collectionForTopic(kind), id)
		case "profile-requests":
			return true
		default:
			return false
		}
	}
	if !activePlayer(auth) {
		return false
	}

	switch kind {
	case "profile":
		return auth.Id == id
	case "participant":
		return participantBelongsToProfile(app, id, auth.Id)
	case "game":
		return playerParticipatesInGame(app, id, auth.Id)
	case "room":
		return playerCanReadRoom(app, id, auth.Id)
	default:
		return false
	}
}

func splitTopic(topic string) (string, string, bool) {
	parts := strings.Split(topic, ":")
	if len(parts) == 2 && (parts[0] == "room" || parts[0] == "profile") && parts[1] != "" {
		return parts[0], parts[1], true
	}
	if len(parts) == 2 && parts[0] == "profile-requests" && parts[1] == "game-masters" {
		return "profile-requests", parts[1], true
	}
	if len(parts) == 3 && parts[0] == "participant" && parts[2] == "private" && parts[1] != "" {
		return parts[0], parts[1], true
	}
	if len(parts) == 3 && parts[0] == "game" && parts[1] != "" {
		if parts[2] == "public" {
			return "game", parts[1], true
		}
		if parts[2] == "game-masters" {
			return "game-master", parts[1], true
		}
	}
	return "", "", false
}

func splitProfileRequestTopic(topic string) (string, string, bool) {
	parts := strings.Split(topic, ":")
	if len(parts) != 3 || parts[0] != "profile-request" || parts[1] == "" || len(parts[2]) != 64 {
		return "", "", false
	}
	return parts[1], parts[2], true
}

func profileRequestCapabilityMatches(app core.App, requestID, capability string) bool {
	record, err := app.FindRecordById("profile_requests", requestID)
	if err != nil {
		return false
	}
	expected := record.GetString("secret_hash")
	if len(expected) != len(capability) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(capability)) == 1
}

func collectionForTopic(kind string) string {
	switch kind {
	case "game", "game-master":
		return "games"
	case "participant":
		return "participants"
	case "room":
		return "chat_rooms"
	case "profile":
		return "player_profiles"
	default:
		return ""
	}
}

func recordExists(app core.App, collection, id string) bool {
	if collection == "" {
		return false
	}
	_, err := app.FindRecordById(collection, id)
	return err == nil
}

func activeGameMaster(auth *core.Record) bool {
	return auth != nil && auth.Collection().Name == "game_masters" && auth.GetBool("active")
}

func activePlayer(auth *core.Record) bool {
	return auth != nil && auth.Collection().Name == "player_profiles" && auth.GetBool("active")
}

func participantBelongsToProfile(app core.App, participantID, profileID string) bool {
	record, err := app.FindRecordById("participants", participantID)
	if err != nil || record.GetString("profile") != profileID {
		return false
	}
	status := record.GetString("status")
	return status != "kicked" && status != "left"
}

func playerParticipatesInGame(app core.App, gameID, profileID string) bool {
	records, err := app.FindRecordsByFilter("participants",
		"game = {:game} && profile = {:profile} && status != 'kicked' && status != 'left'",
		"",
		1,
		0,
		dbx.Params{"game": gameID, "profile": profileID},
	)
	return err == nil && len(records) == 1
}

func playerCanReadRoom(app core.App, roomID, profileID string) bool {
	records, err := app.FindRecordsByFilter("chat_memberships",
		"room = {:room} && participant.profile = {:profile} && ((left_at = '' && participant.status != 'kicked' && participant.status != 'left') || historical_access = true)",
		"",
		1,
		0,
		dbx.Params{"room": roomID, "profile": profileID},
	)
	return err == nil && len(records) == 1
}

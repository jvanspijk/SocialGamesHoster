package auth

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

const (
	GameMastersCollection    = "game_masters"
	PlayerProfilesCollection = "player_profiles"
)

func RequireGameMaster(event *core.RequestEvent) error {
	if event.Auth == nil || event.Auth.Collection().Name != GameMastersCollection || !event.Auth.GetBool("active") {
		return event.JSON(http.StatusUnauthorized, map[string]any{
			"code":    "auth.required",
			"message": "A game-master account is required.",
		})
	}
	return event.Next()
}

func RequireOwner(event *core.RequestEvent) error {
	if event.Auth == nil || event.Auth.Collection().Name != GameMastersCollection || !event.Auth.GetBool("active") || !event.Auth.GetBool("is_owner") {
		return event.JSON(http.StatusForbidden, map[string]any{
			"code":    "auth.owner_required",
			"message": "The owner account is required.",
		})
	}
	return event.Next()
}

func RequirePlayer(event *core.RequestEvent) error {
	if event.Auth == nil || event.Auth.Collection().Name != PlayerProfilesCollection || !event.Auth.GetBool("active") {
		return event.JSON(http.StatusUnauthorized, map[string]any{
			"code":    "auth.required",
			"message": "An approved player profile is required.",
		})
	}
	return event.Next()
}

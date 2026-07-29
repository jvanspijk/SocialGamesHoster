// Package actors owns application actor vocabulary, classification, and
// account-level route authorization.
package actors

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"
)

const (
	GameMastersCollection    = "game_masters"
	PlayerProfilesCollection = "player_profiles"
)

type authorizationError struct {
	status  int
	code    string
	message string
}

func IsGameMaster(actor *core.Record) bool {
	return actor != nil && actor.Collection().Name == GameMastersCollection
}

func IsPlayer(actor *core.Record) bool {
	return actor != nil && actor.Collection().Name == PlayerProfilesCollection
}

func IsActiveGameMaster(actor *core.Record) bool {
	return IsGameMaster(actor) && actor.GetBool("active")
}

func IsActivePlayer(actor *core.Record) bool {
	return IsPlayer(actor) && actor.GetBool("active")
}

func IsOwner(actor *core.Record) bool {
	return IsActiveGameMaster(actor) && actor.GetBool("is_owner")
}

func RequireGameMaster(event *core.RequestEvent) error {
	if authError := requireGameMaster(event.Auth); authError != nil {
		return writeAuthorizationError(event, authError)
	}
	return event.Next()
}

func RequireOwner(event *core.RequestEvent) error {
	if authError := requireOwner(event.Auth); authError != nil {
		return writeAuthorizationError(event, authError)
	}
	return event.Next()
}

func RequirePlayer(event *core.RequestEvent) error {
	if authError := requirePlayer(event.Auth); authError != nil {
		return writeAuthorizationError(event, authError)
	}
	return event.Next()
}

func requireGameMaster(actor *core.Record) *authorizationError {
	if IsActiveGameMaster(actor) {
		return nil
	}
	return &authorizationError{
		status:  http.StatusUnauthorized,
		code:    "auth.required",
		message: "A game-master account is required.",
	}
}

func requireOwner(actor *core.Record) *authorizationError {
	if IsOwner(actor) {
		return nil
	}
	return &authorizationError{
		status:  http.StatusForbidden,
		code:    "auth.owner_required",
		message: "The owner account is required.",
	}
}

func requirePlayer(actor *core.Record) *authorizationError {
	if IsActivePlayer(actor) {
		return nil
	}
	return &authorizationError{
		status:  http.StatusUnauthorized,
		code:    "auth.required",
		message: "An approved player profile is required.",
	}
}

func writeAuthorizationError(event *core.RequestEvent, authError *authorizationError) error {
	return event.JSON(authError.status, map[string]any{
		"code":    authError.code,
		"message": authError.message,
	})
}

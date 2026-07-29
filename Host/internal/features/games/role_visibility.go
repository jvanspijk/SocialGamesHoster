package games

import (
	"net/http"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type roleVisibilityRequest struct {
	RolesVisible bool `json:"rolesVisible"`
}

func setRoleVisibility(event *core.RequestEvent) error {
	game, err := findGame(event)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	status := gamepolicy.GameStatus(game.GetString("status"))
	if status == gamepolicy.GameDraft || gamepolicy.IsArchived(status) {
		return httpx.WriteError(event, result.Conflict("game.role_visibility_not_allowed", "Role visibility can only change in an active game."))
	}
	var request roleVisibilityRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("game.role_visibility_invalid", "The role visibility update could not be read.", nil))
	}
	if request.RolesVisible == game.GetBool("roles_visible") {
		return event.JSON(http.StatusOK, projectRoleVisibility(game))
	}
	if request.RolesVisible {
		participants, err := currentParticipants(event.App, game.Id)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		if len(participants) == 0 {
			return httpx.WriteError(event, result.Conflict("game.assignments_incomplete", "Add players and assign every active player a role first."))
		}
		definition, err := snapshot(game)
		if err != nil {
			return httpx.WriteError(event, result.Internal(err))
		}
		validRoles := make(map[string]bool, len(definition.Roles))
		for _, role := range definition.Roles {
			validRoles[role.ID] = true
		}
		for _, participant := range participants {
			if !validRoles[participant.GetString("role_key")] {
				return httpx.WriteError(event, result.Conflict("game.assignments_incomplete", "Assign every active player a valid role first."))
			}
		}
	}
	game.Set("roles_visible", request.RolesVisible)
	game.Set("role_visibility_revision", game.GetInt("role_visibility_revision")+1)
	incrementRevision(game)
	if err := event.App.Save(game); err != nil {
		return httpx.WriteError(event, result.Internal(err))
	}
	action := "game.roles_hidden"
	if request.RolesVisible {
		action = "game.roles_available"
	}
	payload := projectRoleVisibility(game)
	_ = audit(event.App, event.Auth, game.Id, action, "game", game.Id, payload, event.Get(httpx.TraceIDKey))
	publishGame(event.App, game, "game.role_visibility_changed", payload)
	return event.JSON(http.StatusOK, payload)
}

func projectRoleVisibility(game *core.Record) map[string]any {
	return map[string]any{
		"rolesVisible":           game.GetBool("roles_visible"),
		"roleVisibilityRevision": game.GetInt("role_visibility_revision"),
		"revision":               game.GetInt("revision"),
	}
}

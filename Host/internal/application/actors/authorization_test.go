package actors

import (
	"net/http"
	"testing"

	"github.com/pocketbase/pocketbase/core"
)

func TestActorClassification(t *testing.T) {
	gameMasters := core.NewAuthCollection(GameMastersCollection)
	players := core.NewAuthCollection(PlayerProfilesCollection)
	other := core.NewAuthCollection("other")

	record := func(collection *core.Collection, active, owner bool) *core.Record {
		actor := core.NewRecord(collection)
		actor.Set("active", active)
		actor.Set("is_owner", owner)
		return actor
	}

	tests := []struct {
		name             string
		actor            *core.Record
		gameMaster       bool
		player           bool
		activeGameMaster bool
		activePlayer     bool
		owner            bool
	}{
		{name: "missing auth"},
		{name: "wrong collection", actor: record(other, true, true)},
		{name: "inactive game master", actor: record(gameMasters, false, true), gameMaster: true},
		{name: "active game master", actor: record(gameMasters, true, false), gameMaster: true, activeGameMaster: true},
		{name: "owner", actor: record(gameMasters, true, true), gameMaster: true, activeGameMaster: true, owner: true},
		{name: "inactive owner flag", actor: record(gameMasters, false, true), gameMaster: true},
		{name: "inactive player", actor: record(players, false, false), player: true},
		{name: "active player", actor: record(players, true, false), player: true, activePlayer: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsGameMaster(test.actor); got != test.gameMaster {
				t.Errorf("IsGameMaster() = %t, want %t", got, test.gameMaster)
			}
			if got := IsPlayer(test.actor); got != test.player {
				t.Errorf("IsPlayer() = %t, want %t", got, test.player)
			}
			if got := IsActiveGameMaster(test.actor); got != test.activeGameMaster {
				t.Errorf("IsActiveGameMaster() = %t, want %t", got, test.activeGameMaster)
			}
			if got := IsActivePlayer(test.actor); got != test.activePlayer {
				t.Errorf("IsActivePlayer() = %t, want %t", got, test.activePlayer)
			}
			if got := IsOwner(test.actor); got != test.owner {
				t.Errorf("IsOwner() = %t, want %t", got, test.owner)
			}
		})
	}
}

func TestRouteGuardPolicy(t *testing.T) {
	gameMasters := core.NewAuthCollection(GameMastersCollection)
	players := core.NewAuthCollection(PlayerProfilesCollection)
	other := core.NewAuthCollection("other")

	record := func(collection *core.Collection, active, owner bool) *core.Record {
		actor := core.NewRecord(collection)
		actor.Set("active", active)
		actor.Set("is_owner", owner)
		return actor
	}
	unauthorizedGameMaster := &authorizationError{http.StatusUnauthorized, "auth.required", "A game-master account is required."}
	forbiddenOwner := &authorizationError{http.StatusForbidden, "auth.owner_required", "The owner account is required."}
	unauthorizedPlayer := &authorizationError{http.StatusUnauthorized, "auth.required", "An approved player profile is required."}

	tests := []struct {
		name  string
		guard func(*core.Record) *authorizationError
		actor *core.Record
		want  *authorizationError
	}{
		{name: "game master missing auth", guard: requireGameMaster, want: unauthorizedGameMaster},
		{name: "game master wrong collection", guard: requireGameMaster, actor: record(players, true, false), want: unauthorizedGameMaster},
		{name: "game master inactive", guard: requireGameMaster, actor: record(gameMasters, false, false), want: unauthorizedGameMaster},
		{name: "game master active", guard: requireGameMaster, actor: record(gameMasters, true, false)},
		{name: "owner missing auth", guard: requireOwner, want: forbiddenOwner},
		{name: "owner wrong collection", guard: requireOwner, actor: record(other, true, true), want: forbiddenOwner},
		{name: "owner inactive", guard: requireOwner, actor: record(gameMasters, false, true), want: forbiddenOwner},
		{name: "owner non-owner game master", guard: requireOwner, actor: record(gameMasters, true, false), want: forbiddenOwner},
		{name: "owner active", guard: requireOwner, actor: record(gameMasters, true, true)},
		{name: "player missing auth", guard: requirePlayer, want: unauthorizedPlayer},
		{name: "player wrong collection", guard: requirePlayer, actor: record(gameMasters, true, false), want: unauthorizedPlayer},
		{name: "player inactive", guard: requirePlayer, actor: record(players, false, false), want: unauthorizedPlayer},
		{name: "player active", guard: requirePlayer, actor: record(players, true, false)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.guard(test.actor)
			if got == nil || test.want == nil {
				if got != test.want {
					t.Fatalf("guard error = %#v, want %#v", got, test.want)
				}
				return
			}
			if *got != *test.want {
				t.Errorf("guard error = %#v, want %#v", got, test.want)
			}
		})
	}
}

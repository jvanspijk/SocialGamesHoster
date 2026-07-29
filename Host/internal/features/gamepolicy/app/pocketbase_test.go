package gamepolicyapp

import (
	"testing"

	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
)

func TestGameMutationErrorPreservesArchivedConflictContract(t *testing.T) {
	game := core.NewRecord(&core.Collection{})
	for _, status := range []gamepolicy.GameStatus{
		gamepolicy.GameDraft,
		gamepolicy.GameLobby,
		gamepolicy.GameRunning,
		gamepolicy.GamePaused,
		gamepolicy.GameReview,
		gamepolicy.GameArchived,
	} {
		game.Set("status", status)
		appError := GameMutationError(game)
		if status != gamepolicy.GameArchived {
			if appError != nil {
				t.Fatalf("%q unexpectedly rejected: %#v", status, appError)
			}
			continue
		}
		if appError == nil ||
			appError.Code != "game.archived_immutable" ||
			appError.Message != "Archived games cannot be changed." ||
			appError.Status != 409 {
			t.Fatalf("unexpected archived conflict: %#v", appError)
		}
	}
}

package gamepolicy

import "testing"

func TestParticipantMembershipStatuses(t *testing.T) {
	tests := []struct {
		status  ParticipantStatus
		current bool
		active  bool
	}{
		{status: ParticipantActive, current: true, active: true},
		{status: ParticipantEliminated, current: true, active: false},
		{status: ParticipantKicked, current: false, active: false},
		{status: ParticipantLeft, current: false, active: false},
	}
	for _, test := range tests {
		t.Run(string(test.status), func(t *testing.T) {
			if got := IsCurrentMember(test.status); got != test.current {
				t.Fatalf("IsCurrentMember(%q) = %t, want %t", test.status, got, test.current)
			}
			if got := IsActivePlayer(test.status); got != test.active {
				t.Fatalf("IsActivePlayer(%q) = %t, want %t", test.status, got, test.active)
			}
		})
	}
}

func TestArchivedGameInvariant(t *testing.T) {
	tests := []struct {
		status   GameStatus
		archived bool
	}{
		{status: GameDraft},
		{status: GameLobby},
		{status: GameRunning},
		{status: GamePaused},
		{status: GameReview},
		{status: GameArchived, archived: true},
	}
	for _, test := range tests {
		t.Run(string(test.status), func(t *testing.T) {
			if got := IsArchived(test.status); got != test.archived {
				t.Fatalf("IsArchived(%q) = %t, want %t", test.status, got, test.archived)
			}
		})
	}
}

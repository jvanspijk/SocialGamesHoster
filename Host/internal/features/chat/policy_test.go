package chat

import (
	"testing"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
)

func TestPolicyOrderHonorsMembershipAndManualLock(t *testing.T) {
	base := rulesets.RoomPermission{Visible: true, Readable: true, Sendable: true}
	overrideSend := true
	effective := EffectivePolicy(base, &rulesets.PartialRoomPermission{Sendable: &overrideSend}, ParticipantState{
		IsMember: true,
		IsActive: true,
	}, RoomState{ManuallyLocked: true})
	if effective.Sendable {
		t.Fatal("manual lock must win over phase policy")
	}
}

func TestAnonymousSenderProjection(t *testing.T) {
	sender := Sender{ProfileName: "Secret Name", SeatNumber: 7}
	if label := SenderLabel(sender, rulesets.SenderSeatNumber); label != "Player 7" {
		t.Fatalf("unexpected label: %s", label)
	}
}

package chat

import "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"

type ParticipantState struct {
	IsMember       bool
	IsActive       bool
	HistoricalRead bool
}

type RoomState struct {
	ManuallyLocked           bool
	ManualVisibilityOverride string
}

func EffectivePolicy(base rulesets.RoomPermission, override *rulesets.PartialRoomPermission, participant ParticipantState, room RoomState) rulesets.RoomPermission {
	effective := rulesets.ApplyRoomOverride(base, override)
	if !participant.IsMember {
		effective.Visible = false
		effective.Readable = participant.HistoricalRead
		effective.Sendable = false
	}
	if !participant.IsActive {
		effective.Sendable = false
	}
	if room.ManuallyLocked {
		effective.Sendable = false
	}
	switch room.ManualVisibilityOverride {
	case "visible":
		effective.Visible = participant.IsMember
	case "hidden":
		effective.Visible = false
		effective.Sendable = false
	}
	return effective
}

type Sender struct {
	ProfileName string
	GameAlias   string
	SeatNumber  int
	RoleLabel   string
	TeamLabel   string
}

func SenderLabel(sender Sender, display rulesets.SenderDisplay) string {
	switch display {
	case rulesets.SenderGameAlias:
		if sender.GameAlias != "" {
			return sender.GameAlias
		}
		return sender.ProfileName
	case rulesets.SenderSeatNumber:
		return "Player " + itoa(sender.SeatNumber)
	case rulesets.SenderRoleLabel:
		return sender.RoleLabel
	case rulesets.SenderTeamLabel:
		return sender.TeamLabel
	default:
		return sender.ProfileName
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := make([]byte, 0, 3)
	for value > 0 {
		digits = append(digits, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

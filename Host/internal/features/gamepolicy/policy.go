package gamepolicy

import "errors"

type GameStatus string

const (
	GameDraft    GameStatus = "draft"
	GameLobby    GameStatus = "lobby"
	GameRunning  GameStatus = "running"
	GamePaused   GameStatus = "paused"
	GameReview   GameStatus = "review"
	GameArchived GameStatus = "archived"
)

type ParticipantStatus string

const (
	ParticipantActive     ParticipantStatus = "active"
	ParticipantEliminated ParticipantStatus = "eliminated"
	ParticipantKicked     ParticipantStatus = "kicked"
	ParticipantLeft       ParticipantStatus = "left"
)

var ErrArchivedImmutable = errors.New("archived games are immutable")

// IsCurrentMember reports whether a participant still belongs to the game.
// Eliminated participants remain members even though they are not eligible for
// capabilities that require a player who is active in play.
func IsCurrentMember(status ParticipantStatus) bool {
	return status == ParticipantActive || status == ParticipantEliminated
}

func IsActivePlayer(status ParticipantStatus) bool {
	return status == ParticipantActive
}

func IsArchived(status GameStatus) bool {
	return status == GameArchived
}

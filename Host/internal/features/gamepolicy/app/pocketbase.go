package gamepolicyapp

import (
	"errors"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

var errCurrentParticipantNotFound = errors.New("current participant not found")

const (
	CurrentParticipantStatusFilter        = "status != 'kicked' && status != 'left'"
	CurrentRelatedParticipantStatusFilter = "participant.status != 'kicked' && participant.status != 'left'"

	currentParticipantByGameAndProfileFilter = "game = {:game} && profile = {:profile} && " + CurrentParticipantStatusFilter
	currentParticipantsByGameFilter          = "game = {:game} && " + CurrentParticipantStatusFilter
	currentParticipantsByProfileFilter       = "profile = {:profile} && " + CurrentParticipantStatusFilter

	RoomReadableByCurrentOrHistoricalParticipantFilter = "room = {:room} && participant.profile = {:profile} && ((left_at = '' && " + CurrentRelatedParticipantStatusFilter + ") || historical_access = true)"
)

func CurrentParticipantByGameAndProfile(app core.App, gameID, profileID string) (*core.Record, error) {
	records, err := app.FindRecordsByFilter(
		"participants",
		currentParticipantByGameAndProfileFilter,
		"",
		1,
		0,
		dbx.Params{"game": gameID, "profile": profileID},
	)
	if err != nil {
		return nil, err
	}
	if len(records) != 1 {
		return nil, errCurrentParticipantNotFound
	}
	return records[0], nil
}

func CurrentParticipantsByGame(app core.App, gameID string) ([]*core.Record, error) {
	return app.FindRecordsByFilter(
		"participants",
		currentParticipantsByGameFilter,
		"seat_number",
		30,
		0,
		dbx.Params{"game": gameID},
	)
}

func CurrentParticipantsByProfile(app core.App, profileID string) ([]*core.Record, error) {
	return app.FindRecordsByFilter(
		"participants",
		currentParticipantsByProfileFilter,
		"",
		1000,
		0,
		dbx.Params{"profile": profileID},
	)
}

func ParticipantBelongsToProfile(app core.App, participantID, profileID string) bool {
	record, err := app.FindRecordById("participants", participantID)
	return err == nil &&
		record.GetString("profile") == profileID &&
		gamepolicy.IsCurrentMember(gamepolicy.ParticipantStatus(record.GetString("status")))
}

func ProfileParticipatesInGame(app core.App, gameID, profileID string) bool {
	_, err := CurrentParticipantByGameAndProfile(app, gameID, profileID)
	return err == nil
}

// GameMutationError returns the stable general-purpose mutation error. Features
// with a more specific failed capability should use gamepolicy.IsArchived and
// retain their capability-specific public error.
func GameMutationError(game *core.Record) *result.AppError {
	if game == nil || !gamepolicy.IsArchived(gamepolicy.GameStatus(game.GetString("status"))) {
		return nil
	}
	value := result.Conflict("game.archived_immutable", "Archived games cannot be changed.")
	return &value
}

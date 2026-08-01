package timer

import (
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
	applicationaudit "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/audit"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/abilities"
	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/realtime"
)

type Service struct {
	app      core.App
	mu       sync.Mutex
	timer    *time.Timer
	gameID   string
	revision int
}

// auditRecord is replaced only by package tests to exercise transaction
// rollback when durable auditing is unavailable.
var auditRecord = applicationaudit.Record

func NewService(app core.App) *Service {
	return &Service{app: app}
}

func (service *Service) Reconcile() {
	records, err := service.app.FindRecordsByFilter("games",
		"(status = 'running' || status = 'paused') && (timer_state = 'running' || (timer_state = 'completed' && ability_phase_locked_at = ''))", "", 1, 0)
	if err != nil || len(records) == 0 {
		return
	}
	game := records[0]
	state := Reconcile(stateFromRecord(game), time.Now().UTC())
	if state.Status == Completed {
		now := time.Now().UTC()
		if completed, err := service.completeTransaction(game.Id, state.Revision, now); err == nil && completed {
			game, _ = service.app.FindRecordById("games", game.Id)
			service.Publish(game, "timer.completed", Project(state, time.Now().UTC()))
		}
		return
	}
	service.Schedule(game.Id, state)
}

func (service *Service) Schedule(gameID string, state State) {
	service.mu.Lock()
	defer service.mu.Unlock()
	if service.timer != nil {
		service.timer.Stop()
		service.timer = nil
	}
	service.gameID = gameID
	service.revision = state.Revision
	if state.Status != Running || state.EndsAt == nil {
		return
	}
	delay := time.Until(*state.EndsAt)
	if delay < 0 {
		delay = 0
	}
	revision := state.Revision
	service.timer = time.AfterFunc(delay, func() {
		service.complete(gameID, revision)
	})
}

func (service *Service) complete(gameID string, revision int) {
	service.mu.Lock()
	if service.gameID != gameID || service.revision != revision {
		service.mu.Unlock()
		return
	}
	service.timer = nil
	service.mu.Unlock()

	game, err := service.app.FindRecordById("games", gameID)
	if err != nil {
		return
	}
	state := stateFromRecord(game)
	if state.Status != Running || state.Revision != revision || state.EndsAt == nil {
		return
	}
	now := time.Now().UTC()
	state = Reconcile(state, now)
	if state.Status != Completed {
		service.Schedule(gameID, state)
		return
	}
	completed, err := service.completeTransaction(game.Id, revision, now)
	if err != nil || !completed {
		return
	}
	game, _ = service.app.FindRecordById("games", game.Id)
	service.Publish(game, "timer.completed", Project(state, now))
}

func (service *Service) completeTransaction(gameID string, expectedRevision int, now time.Time) (bool, error) {
	completed := false
	err := service.app.RunInTransaction(func(tx core.App) error {
		game, err := tx.FindRecordById("games", gameID)
		if err != nil {
			return err
		}
		state := Reconcile(stateFromRecord(game), now)
		if state.Status != Completed || state.Revision != expectedRevision {
			return nil
		}
		// A completed record may be reconciled repeatedly; only the transition
		// from a running timer owns completion, revision, audit, and publication.
		if game.GetString("timer_state") != string(Running) {
			return nil
		}
		saveState(game, state)
		if _, err := abilities.FinalizePhase(tx, game, now); err != nil {
			return err
		}
		game.Set("revision", game.GetInt("revision")+1)
		if err := tx.Save(game); err != nil {
			return err
		}
		if err := auditRecord(tx, nil, game.Id, "timer.completed", "game", game.Id, nil, nil); err != nil {
			return err
		}
		completed = true
		return nil
	})
	return completed, err
}

func (service *Service) Publish(game *core.Record, kind string, projection *Projection) {
	_ = realtime.Publish(service.app, "game:"+game.Id+":public", realtime.Event[*Projection]{
		EventID: realtime.NewEventID(), GameID: game.Id, Revision: game.GetInt("revision"),
		Kind: kind, Payload: projection,
	}, func(auth *core.Record) bool {
		if auth == nil || !auth.GetBool("active") {
			return false
		}
		if actorauth.IsGameMaster(auth) {
			return true
		}
		if !actorauth.IsPlayer(auth) {
			return false
		}
		return gamepolicyapp.ProfileParticipatesInGame(service.app, game.Id, auth.Id)
	})
}

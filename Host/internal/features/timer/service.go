package timer

import (
	"sync"
	"time"

	"github.com/pocketbase/pocketbase/core"

	actorauth "github.com/jvanspijk/SocialGamesHoster/Host/internal/application/actors"
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
		saveState(game, state)
		if service.app.Save(game) == nil {
			_, _ = abilities.FinalizePhase(service.app, game.Id, time.Now().UTC())
			if current, findErr := service.app.FindRecordById("games", game.Id); findErr == nil {
				game = current
			}
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
	saveState(game, state)
	if service.app.Save(game) != nil {
		return
	}
	_, _ = abilities.FinalizePhase(service.app, game.Id, now)
	if current, findErr := service.app.FindRecordById("games", game.Id); findErr == nil {
		game = current
	}
	service.Publish(game, "timer.completed", Project(state, now))
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

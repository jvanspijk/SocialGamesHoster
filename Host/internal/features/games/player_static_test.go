package games

import (
	"errors"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/store"
	"github.com/pocketbase/pocketbase/tools/types"
	"sync"
	"sync/atomic"
	"testing"
)

type staticMetadataApp struct {
	core.App
	values *store.Store[string, any]
	calls  atomic.Int32
	fail   atomic.Bool
	assets []*core.Record
}

func (app *staticMetadataApp) Store() *store.Store[string, any] { return app.values }
func (app *staticMetadataApp) FindRecordsByFilter(collection any, filter, sort string, limit, offset int, params ...dbx.Params) ([]*core.Record, error) {
	app.calls.Add(1)
	if app.fail.Load() {
		return nil, errors.New("metadata read failed")
	}
	return app.assets, nil
}
func staticMetadataFixture() (*staticMetadataApp, *core.Record) {
	app := &staticMetadataApp{values: store.New[string, any](nil)}
	for _, key := range []string{"cover", "private-role"} {
		asset := core.NewRecord(core.NewBaseCollection("ruleset_assets"))
		asset.Id = key
		asset.Set("asset_key", key)
		asset.Set("display_name", key)
		asset.Set("kind", "image")
		asset.Set("checksum", "checksum")
		app.assets = append(app.assets, asset)
	}
	game := core.NewRecord(core.NewBaseCollection("games"))
	game.Id = "first-game"
	game.Set("ruleset_version", "frozen-version")
	game.Set("started_at", types.NowDateTime())
	game.Set("ruleset_snapshot", rulesets.DefinitionV1{SchemaVersion: 1, Metadata: rulesets.Metadata{Name: "Frozen"}, Roles: []rulesets.Role{{ID: "role", ImageAssetKey: "private-role"}}})
	return app, game
}
func TestPlayerStaticReusesMetadataButAppliesCurrentVisibility(t *testing.T) {
	app, game := staticMetadataFixture()
	first, err := playerStaticForGame(app, game)
	if err != nil {
		t.Fatal(err)
	}
	for _, visible := range []bool{false, true, false, true} {
		data, err := playerStaticForGame(app, game)
		if err != nil {
			t.Fatal(err)
		}
		assets := data.assetsForPlayer(visible)
		want := 1
		if visible {
			want = 2
		}
		if len(assets) != want {
			t.Fatalf("visibility %t: assets=%d, want %d", visible, len(assets), want)
		}
		if assets[0].Preview != "/api/app/v1/ruleset-assets/cover" || assets[0].Checksum != "checksum" {
			t.Fatal("asset metadata contract changed")
		}
		assets[0].DisplayName = "response-local change"
	}
	if app.calls.Load() != 1 {
		t.Fatalf("metadata loaded %d times", app.calls.Load())
	}
	if first.assetsForPlayer(true)[0].DisplayName != "cover" {
		t.Fatal("response mutated cached metadata")
	}
	if first.definition.Metadata.Name != "Frozen" {
		t.Fatal("definition not retained")
	}
}
func TestPlayerStaticRebuildsForAnotherGameAndHostRestart(t *testing.T) {
	app, game := staticMetadataFixture()
	if _, err := playerStaticForGame(app, game); err != nil {
		t.Fatal(err)
	}
	game.Set("status", "paused")
	if _, err := playerStaticForGame(app, game); err != nil {
		t.Fatal(err)
	}
	if app.calls.Load() != 1 {
		t.Fatal("pause caused reload")
	}
	game.Id = "second-game"
	game.Set("ruleset_snapshot", rulesets.DefinitionV1{SchemaVersion: 1, Metadata: rulesets.Metadata{Name: "Second"}})
	data, err := playerStaticForGame(app, game)
	if err != nil {
		t.Fatal(err)
	}
	if app.calls.Load() != 2 || data.definition.Metadata.Name != "Second" {
		t.Fatal("another game reused stale metadata")
	}
	resumed := &staticMetadataApp{values: store.New[string, any](nil), assets: app.assets}
	data, err = playerStaticForGame(resumed, game)
	if err != nil {
		t.Fatal(err)
	}
	if resumed.calls.Load() != 1 || data.definition.Metadata.Name != "Second" {
		t.Fatal("resumed game did not rebuild")
	}
}
func TestPlayerStaticDoesNotRetainPreStartDataOrLoadErrors(t *testing.T) {
	app, game := staticMetadataFixture()
	game.Set("started_at", nil)
	for i := 0; i < 2; i++ {
		if _, err := playerStaticForGame(app, game); err != nil {
			t.Fatal(err)
		}
	}
	if app.calls.Load() != 2 {
		t.Fatal("pre-start data retained")
	}
	game.Set("started_at", types.NowDateTime())
	app.fail.Store(true)
	if _, err := playerStaticForGame(app, game); err == nil {
		t.Fatal("failed metadata load succeeded")
	}
	app.fail.Store(false)
	if _, err := playerStaticForGame(app, game); err != nil {
		t.Fatal(err)
	}
	if app.calls.Load() != 4 {
		t.Fatal("failed metadata load was retained")
	}
}
func TestPlayerStaticSharesOneLoadAcrossConcurrentSnapshots(t *testing.T) {
	app, game := staticMetadataFixture()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			data, err := playerStaticForGame(app, game)
			if err != nil {
				t.Error(err)
				return
			}
			if len(data.assetsForPlayer(true)) != 2 {
				t.Error("incomplete metadata")
			}
		}()
	}
	wg.Wait()
	if app.calls.Load() != 1 {
		t.Fatalf("concurrent snapshots loaded metadata %d times", app.calls.Load())
	}
}

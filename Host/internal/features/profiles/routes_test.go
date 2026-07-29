package profiles

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/pocketbase/pocketbase/core"

	_ "github.com/jvanspijk/SocialGamesHoster/Host/migrations"
)

func TestClassifyProfileRequest(t *testing.T) {
	profile := &core.Record{}
	wrappedNotFound := fmt.Errorf("wrapped not found: %w", sql.ErrNoRows)

	tests := []struct {
		name        string
		existing    *core.Record
		lookupErr   error
		requestType string
		profile     *core.Record
		wantErr     bool
	}{
		{name: "existing profile", existing: profile, requestType: "recover", profile: profile},
		{name: "missing profile", lookupErr: sql.ErrNoRows, requestType: "new"},
		{name: "wrapped missing profile", lookupErr: wrappedNotFound, requestType: "new"},
		{name: "unexpected lookup error", lookupErr: errors.New("database unavailable"), wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestType, gotProfile, err := classifyProfileRequest(test.existing, test.lookupErr)
			if (err != nil) != test.wantErr {
				t.Fatalf("error = %v, want error = %t", err, test.wantErr)
			}
			if requestType != test.requestType {
				t.Fatalf("request type = %q, want %q", requestType, test.requestType)
			}
			if gotProfile != test.profile {
				t.Fatalf("profile = %p, want %p", gotProfile, test.profile)
			}
		})
	}
}

func TestSyncLiveParticipantNamesUpdatesOnlyLiveGames(t *testing.T) {
	if err := os.MkdirAll(".testdata", 0o700); err != nil {
		t.Fatal(err)
	}
	dataDir, err := os.MkdirTemp(".testdata", "profiles-routes-")
	if err != nil {
		t.Fatal(err)
	}
	dataDir, err = filepath.Abs(dataDir)
	if err != nil {
		t.Fatal(err)
	}
	app := core.NewBaseApp(core.BaseAppConfig{DataDir: dataDir, EncryptionEnv: "sgh_test_encryption"})
	if err := app.Bootstrap(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		app.ResetBootstrapState()
		_ = os.RemoveAll(dataDir)
	})
	if err := app.RunAllMigrations(); err != nil {
		t.Fatal(err)
	}
	gameMasters, _ := app.FindCollectionByNameOrId("game_masters")
	gameMaster := core.NewRecord(gameMasters)
	gameMaster.Set("username", "host")
	gameMaster.Set("display_name", "Host")
	gameMaster.Set("is_owner", true)
	gameMaster.Set("active", true)
	gameMaster.SetPassword("secret-password")
	if err := app.Save(gameMaster); err != nil {
		t.Fatal(err)
	}

	profiles, _ := app.FindCollectionByNameOrId("player_profiles")
	profile := core.NewRecord(profiles)
	profile.Set("display_name", "New Name")
	profile.Set("normalized_name", "new name")
	profile.Set("active", true)
	profile.SetPassword("secret-password")
	if err := app.Save(profile); err != nil {
		t.Fatal(err)
	}
	rulesets, _ := app.FindCollectionByNameOrId("rulesets")
	ruleset := core.NewRecord(rulesets)
	ruleset.Set("slug", "test")
	ruleset.Set("name", "Test")
	ruleset.Set("created_by", gameMaster.Id)
	if err := app.Save(ruleset); err != nil {
		t.Fatal(err)
	}
	ruleVersions, _ := app.FindCollectionByNameOrId("ruleset_versions")
	ruleVersion := core.NewRecord(ruleVersions)
	ruleVersion.Set("ruleset", ruleset.Id)
	ruleVersion.Set("version_number", 1)
	ruleVersion.Set("state", "published")
	ruleVersion.Set("schema_version", 1)
	ruleVersion.Set("definition", map[string]any{"test": true})
	ruleVersion.Set("created_by", gameMaster.Id)
	if err := app.Save(ruleVersion); err != nil {
		t.Fatal(err)
	}

	games, _ := app.FindCollectionByNameOrId("games")
	liveGame := core.NewRecord(games)
	liveGame.Set("name", "Live")
	liveGame.Set("status", "running")
	liveGame.Set("ruleset_version", ruleVersion.Id)
	liveGame.Set("ruleset_snapshot", map[string]any{"test": true})
	liveGame.Set("timer_state", "inactive")
	liveGame.Set("created_by", gameMaster.Id)
	if err := app.Save(liveGame); err != nil {
		t.Fatal(err)
	}
	archivedGame := core.NewRecord(games)
	archivedGame.Set("name", "Archived")
	archivedGame.Set("status", "archived")
	archivedGame.Set("ruleset_version", ruleVersion.Id)
	archivedGame.Set("ruleset_snapshot", map[string]any{"test": true})
	archivedGame.Set("timer_state", "inactive")
	archivedGame.Set("created_by", gameMaster.Id)
	if err := app.Save(archivedGame); err != nil {
		t.Fatal(err)
	}

	participants, _ := app.FindCollectionByNameOrId("participants")
	for _, game := range []*core.Record{liveGame, archivedGame} {
		participant := core.NewRecord(participants)
		participant.Set("game", game.Id)
		participant.Set("profile", profile.Id)
		participant.Set("display_name_snapshot", "Old Name")
		participant.Set("status", "active")
		participant.Set("seat_number", 1)
		participant.Set("outcome", "unset")
		participant.Set("joined_at", time.Now().UTC())
		if err := app.Save(participant); err != nil {
			t.Fatal(err)
		}
	}

	if err := syncLiveParticipantNames(app, profile); err != nil {
		t.Fatal(err)
	}
	for _, testCase := range []struct {
		game *core.Record
		name string
	}{
		{game: liveGame, name: "New Name"},
		{game: archivedGame, name: "Old Name"},
	} {
		updated, err := app.FindRecordsByFilter("participants", "profile = {:profile} && game = {:game}", "", 1, 0,
			map[string]any{"profile": profile.Id, "game": testCase.game.Id})
		if err != nil {
			t.Fatal(err)
		}
		if len(updated) != 1 || updated[0].GetString("display_name_snapshot") != testCase.name {
			t.Fatalf("%s participant name = %q", testCase.game.GetString("status"), updated[0].GetString("display_name_snapshot"))
		}
	}
}

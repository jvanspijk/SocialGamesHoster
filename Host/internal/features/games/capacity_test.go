package games

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	gamepolicyapp "github.com/jvanspijk/SocialGamesHoster/Host/internal/features/gamepolicy/app"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/testutil"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

func TestCapacityIncludesEveryPlayerAndRejectsAnExtraJoin(t *testing.T) {
	app := testutil.NewPocketBaseApp(t)
	// Persistence fixtures isolate roster and join policy from identity setup.
	_, err := app.DB().NewQuery(`INSERT INTO games (id,name,status,joining_open,ruleset_snapshot)
		VALUES ('capacitygame001','Capacity','lobby',true,
		'{"schemaVersion":1,"metadata":{"name":"Capacity","minPlayers":1,"maxPlayers":255}}')`).Execute()
	if err != nil {
		t.Fatal(err)
	}
	for number := 1; number <= 255; number++ {
		_, err = app.DB().NewQuery(`INSERT INTO participants
			(id,game,profile,display_name_snapshot,player_number,status,outcome)
			VALUES ({:id},'capacitygame001',{:profile},'Player',{:number},'active','unset')`).Bind(dbx.Params{
			"id": fmt.Sprintf("participant%04d", number), "profile": fmt.Sprintf("profile%08d", number), "number": number,
		}).Execute()
		if err != nil {
			t.Fatal(err)
		}
	}
	roster, err := gameParticipants(app, "capacitygame001")
	if err != nil || len(roster) != 255 {
		t.Fatalf("full roster: count=%d err=%v", len(roster), err)
	}
	current, err := gamepolicyapp.CurrentParticipantsByGame(app, "capacitygame001")
	if err != nil || len(current) != 255 {
		t.Fatalf("current roster: count=%d err=%v", len(current), err)
	}
	if roster[254].GetInt("player_number") != 255 {
		t.Fatal("last player omitted")
	}
	collection, err := app.FindCollectionByNameOrId("player_profiles")
	if err != nil {
		t.Fatal(err)
	}
	actor := core.NewRecord(collection)
	actor.Id = "extraprofile001"
	request := httptest.NewRequest(http.MethodPost, "/api/app/v1/games/capacitygame001/join", nil)
	request.SetPathValue("id", "capacitygame001")
	response := httptest.NewRecorder()
	event := &core.RequestEvent{App: app, Auth: actor}
	event.Request = request
	event.Response = response
	if err := joinGame(event); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusConflict {
		t.Fatalf("extra join status=%d body=%s", response.Code, response.Body.String())
	}
	var body struct {
		Code string `json:"code"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != "game.full" {
		t.Fatalf("extra join: %s", response.Body.String())
	}
}

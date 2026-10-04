package games

import (
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/features/rulesets"
	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
	"sync"
)

const playerStaticStoreKey = "games.playerStatic"

// These values belong to the frozen game definition, never to a player's role
// or permissions. Keep the cached definition read-only; views project fresh state.
type playerStaticData struct {
	gameID           string
	versionID        string
	definition       rulesets.DefinitionV1
	assets           []playerAssetMetadata
	privateAssetKeys map[string]bool
}

type playerAssetMetadata struct {
	ID                string `json:"id"`
	AssetKey          string `json:"assetKey"`
	Kind              string `json:"kind"`
	DisplayName       string `json:"displayName"`
	AccessibilityText string `json:"accessibilityText"`
	Checksum          string `json:"checksum"`
	Preview           string `json:"preview"`
}

type playerStaticState struct {
	sync.Mutex
	data *playerStaticData
}

func staticState(app core.App) *playerStaticState {
	return app.Store().GetOrSet(playerStaticStoreKey, func() any { return &playerStaticState{} }).(*playerStaticState)
}

func playerStaticForGame(app core.App, game *core.Record) (*playerStaticData, error) {
	// Lobby views still work before Start, without keeping pre-start data.
	if game.GetDateTime("started_at").IsZero() {
		return loadPlayerStaticData(app, game)
	}
	state := staticState(app)
	state.Lock()
	defer state.Unlock()
	if state.data != nil && state.data.gameID == game.Id && state.data.versionID == game.GetString("ruleset_version") {
		return state.data, nil
	}
	// Also handles an active game resumed after a process restart. Failed loads
	// are not retained, and concurrent cold requests share one successful load.
	data, err := loadPlayerStaticData(app, game)
	if err != nil {
		return nil, err
	}
	state.data = data
	return data, nil
}

func loadPlayerStaticData(app core.App, game *core.Record) (*playerStaticData, error) {
	definition, err := snapshot(game)
	if err != nil {
		return nil, err
	}
	records, err := app.FindRecordsByFilter("ruleset_assets", "ruleset_version = {:version} && storage_state = 'ready'", "asset_key", 100, 0, dbx.Params{"version": game.GetString("ruleset_version")})
	if err != nil {
		return nil, err
	}
	data := &playerStaticData{gameID: game.Id, versionID: game.GetString("ruleset_version"), definition: definition, assets: make([]playerAssetMetadata, 0, len(records)), privateAssetKeys: map[string]bool{}}
	for _, key := range privateRoleAssetKeys(definition) {
		data.privateAssetKeys[key] = true
	}
	for _, asset := range records {
		data.assets = append(data.assets, playerAssetMetadata{ID: asset.Id, AssetKey: asset.GetString("asset_key"), Kind: asset.GetString("kind"), DisplayName: asset.GetString("display_name"), AccessibilityText: asset.GetString("accessibility_text"), Checksum: asset.GetString("checksum"), Preview: "/api/app/v1/ruleset-assets/" + asset.Id})
	}
	return data, nil
}

func rememberPlayerStaticData(app core.App, data *playerStaticData) {
	state := staticState(app)
	state.Lock()
	defer state.Unlock()
	state.data = data
}

// Return a separate slice so a response cannot mutate the game-wide metadata.
// Visibility is decided from the current request, not stored in the cache.
func (data *playerStaticData) assetsForPlayer(roleAvailable bool) []playerAssetMetadata {
	assets := make([]playerAssetMetadata, 0, len(data.assets))
	for _, asset := range data.assets {
		if !roleAvailable && data.privateAssetKeys[asset.AssetKey] {
			continue
		}
		assets = append(assets, asset)
	}
	return assets
}

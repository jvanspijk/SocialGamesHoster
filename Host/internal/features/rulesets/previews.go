package rulesets

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"

	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/httpx"
	"github.com/jvanspijk/SocialGamesHoster/Host/internal/platform/result"
)

type previewRequest struct {
	Definition  DefinitionV1 `json:"definition"`
	SessionID   string       `json:"sessionId,omitempty"`
	Mode        string       `json:"mode"`
	RoleID      string       `json:"roleId,omitempty"`
	PhaseID     string       `json:"phaseId,omitempty"`
	PlayerCount int          `json:"playerCount,omitempty"`
	AssetKey    string       `json:"assetKey,omitempty"`
}

func previewRuleset(event *core.RequestEvent) error {
	logical, err := event.App.FindRecordById("rulesets", event.Request.PathValue("id"))
	if err != nil || logical.GetBool("archived") {
		return rulesetNotFound(event)
	}
	var request previewRequest
	if err := event.BindBody(&request); err != nil {
		return httpx.WriteError(event, result.Invalid("ruleset.preview_invalid", "The preview settings could not be read.", nil))
	}
	assets, err := previewAssets(event, logical, request.SessionID)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	response, err := buildRulesetPreview(request, assets)
	if err != nil {
		return httpx.WriteErrorFrom(event, err)
	}
	return event.JSON(http.StatusOK, response)
}

func previewAssets(event *core.RequestEvent, logical *core.Record, sessionID string) (map[string]effectiveAsset, error) {
	if sessionID != "" {
		event.Request.SetPathValue("sessionId", sessionID)
		session, err := ownedEditSession(event)
		if err != nil {
			return nil, err
		}
		assets, err := effectiveSessionAssets(event.App, session)
		if err != nil {
			return nil, err
		}
		touchEditSession(session, time.Now().UTC())
		if err := event.App.Save(session); err != nil {
			return nil, err
		}
		return assets, nil
	}
	version, err := latestSavedVersion(event.App, logical)
	if err != nil {
		return nil, err
	}
	records, err := event.App.FindAllRecords("ruleset_assets", dbx.HashExp{"ruleset_version": version.Id, "storage_state": "ready"})
	if err != nil {
		return nil, err
	}
	assets := make(map[string]effectiveAsset, len(records))
	for _, record := range records {
		assets[record.GetString("asset_key")] = effectiveAssetFromRecord(record, false)
	}
	return assets, nil
}

func buildRulesetPreview(request previewRequest, assets map[string]effectiveAsset) (map[string]any, error) {
	switch request.Mode {
	case "role":
		return rolePreview(request.Definition, assets, request.RoleID)
	case "phases":
		return phasePreview(request.Definition, assets, request.PhaseID), nil
	case "composition":
		return compositionPreview(request.Definition, request.PlayerCount), nil
	case "chat":
		return chatPreview(request.Definition, request.RoleID, request.PhaseID)
	case "media":
		return mediaPreview(request.Definition, assets, request.AssetKey)
	default:
		return nil, result.Invalid("ruleset.preview_mode_invalid", "Choose a preview type.", nil)
	}
}

func rolePreview(definition DefinitionV1, assets map[string]effectiveAsset, roleID string) (map[string]any, error) {
	role, ok := selectedRole(definition, roleID)
	if !ok {
		return map[string]any{"mode": "role", "empty": true, "message": "Add a role to preview the player role card."}, nil
	}
	teamName := "No team"
	for _, team := range definition.Teams {
		if team.ID == role.TeamID {
			teamName = team.Name
			break
		}
	}
	abilities := make([]map[string]any, 0, len(role.AbilityIDs))
	for _, id := range role.AbilityIDs {
		for _, ability := range definition.Abilities {
			if ability.ID == id {
				abilities = append(abilities, map[string]any{"name": ability.Name, "description": ability.Description})
			}
		}
	}
	response := map[string]any{"mode": "role", "role": map[string]any{
		"name": role.Name, "description": role.Description, "teamName": teamName,
		"winCondition": role.WinCondition, "abilities": abilities,
	}}
	if asset, ok := assets[role.ImageAssetKey]; ok {
		response["media"] = previewMediaSummary(asset)
	}
	return response, nil
}

func phasePreview(definition DefinitionV1, assets map[string]effectiveAsset, selectedID string) map[string]any {
	phases := append([]Phase(nil), definition.Phases...)
	sort.SliceStable(phases, func(i, j int) bool {
		if phases[i].Order == phases[j].Order {
			return phases[i].Name < phases[j].Name
		}
		return phases[i].Order < phases[j].Order
	})
	if len(phases) == 0 {
		return map[string]any{"mode": "phases", "empty": true, "message": "This ruleset has no game flow. The game master can run it without phases."}
	}
	items := make([]map[string]any, 0, len(phases))
	for _, phase := range phases {
		item := map[string]any{
			"name": phase.Name, "description": phase.Description, "startsRound": phase.StartsRound,
			"suggestedDurationSeconds": phase.SuggestedDurationSeconds, "selected": phase.ID == selectedID,
		}
		for _, cue := range definition.AudioCues {
			if cue.ID == phase.AudioCueID {
				item["sound"] = cue.Name
				if asset, ok := assets[cue.AssetKey]; ok {
					item["media"] = previewMediaSummary(asset)
				}
			}
		}
		items = append(items, item)
	}
	return map[string]any{"mode": "phases", "phases": items}
}

func compositionPreview(definition DefinitionV1, playerCount int) map[string]any {
	if playerCount == 0 {
		playerCount = definition.Metadata.MinPlayers
	}
	response := map[string]any{"mode": "composition", "playerCount": playerCount}
	if playerCount < 1 || playerCount > 30 {
		response["feasible"] = false
		response["message"] = "Choose a player count between 1 and 30."
		return response
	}
	participants := make([]string, playerCount)
	for index := range participants {
		participants[index] = fmt.Sprintf("seat-%d", index+1)
	}
	assignments, err := RandomizeAssignments(definition, participants, nil, 1)
	if err != nil {
		response["feasible"] = false
		response["message"] = compositionPreviewMessage(err, playerCount)
		return response
	}
	roles := make(map[string]Role, len(definition.Roles))
	teams := make(map[string]string, len(definition.Teams))
	for _, role := range definition.Roles {
		roles[role.ID] = role
	}
	for _, team := range definition.Teams {
		teams[team.ID] = team.Name
	}
	counts := map[string]int{}
	teamByRoleName := map[string]string{}
	for _, assignment := range assignments {
		role := roles[assignment.RoleID]
		counts[role.Name]++
		teamByRoleName[role.Name] = teams[role.TeamID]
	}
	names := make([]string, 0, len(counts))
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	roleCounts := make([]map[string]any, 0, len(names))
	for _, name := range names {
		roleCounts = append(roleCounts, map[string]any{"name": name, "teamName": teamByRoleName[name], "count": counts[name]})
	}
	response["feasible"] = true
	response["message"] = fmt.Sprintf("A valid setup is available for %d players.", playerCount)
	response["roles"] = roleCounts
	return response
}

func compositionPreviewMessage(err error, playerCount int) string {
	message := err.Error()
	switch {
	case strings.Contains(message, "no composition band"):
		return fmt.Sprintf("No player setup covers %d players.", playerCount)
	case strings.Contains(message, "base composition slots total"):
		return fmt.Sprintf("The player setup does not provide exactly %d role places.", playerCount)
	default:
		return fmt.Sprintf("No valid role combination is available for %d players.", playerCount)
	}
}

func chatPreview(definition DefinitionV1, roleID, phaseID string) (map[string]any, error) {
	role, ok := selectedRole(definition, roleID)
	if !ok {
		return map[string]any{"mode": "chat", "empty": true, "message": "Add a role to preview chat availability."}, nil
	}
	phaseName := "No phase"
	for _, phase := range definition.Phases {
		if phase.ID == phaseID {
			phaseName = phase.Name
			break
		}
	}
	override := definition.Chat.PhaseOverrides[phaseID]
	rooms := []map[string]any{}
	if definition.Chat.DefaultPolicy.General != nil {
		rooms = append(rooms, previewRoom("General", "General chat", ApplyRoomOverride(*definition.Chat.DefaultPolicy.General, override.General), true, ChatNormalText))
	}
	if definition.Chat.DefaultPolicy.PlayerDM != nil {
		rooms = append(rooms, previewRoom("Direct messages", "Player messages", ApplyRoomOverride(*definition.Chat.DefaultPolicy.PlayerDM, override.PlayerDM), true, ChatNormalText))
	}
	if base, exists := definition.Chat.DefaultPolicy.Teams[role.TeamID]; exists {
		teamName := "Team chat"
		for _, team := range definition.Teams {
			if team.ID == role.TeamID {
				teamName = team.Name
			}
		}
		var teamOverride *PartialRoomPermission
		if value, exists := override.Teams[role.TeamID]; exists {
			teamOverride = &value
		}
		rooms = append(rooms, previewRoom(teamName, "Team chat", ApplyRoomOverride(base, teamOverride), true, ChatNormalText))
	}
	for _, channel := range definition.Chat.Channels {
		base, phaseOverride := ChatChannelPolicy(channel, phaseID)
		reader := ChatChannelAudienceMatches(channel, role, false)
		sender := reader && ChatChannelAudienceMatches(channel, role, true)
		permission := ApplyRoomOverride(base, phaseOverride)
		if !sender {
			permission.Sendable = false
		}
		rooms = append(rooms, previewRoom(channel.Name, "Custom channel", permission, reader, channel.MessageRestriction))
	}
	return map[string]any{"mode": "chat", "audience": role.Name, "phase": phaseName, "rooms": rooms}, nil
}

func previewRoom(name, kind string, permission RoomPermission, member bool, restriction ChatMessageRestriction) map[string]any {
	if !member {
		permission.Visible = false
		permission.Readable = false
		permission.Sendable = false
	}
	return map[string]any{"name": name, "kind": kind, "visible": permission.Visible, "readable": permission.Readable, "sendable": permission.Sendable, "messageRestriction": restriction}
}

func mediaPreview(definition DefinitionV1, assets map[string]effectiveAsset, key string) (map[string]any, error) {
	if key == "" {
		keys := make([]string, 0, len(assets))
		for assetKey := range assets {
			keys = append(keys, assetKey)
		}
		sort.Slice(keys, func(i, j int) bool { return assets[keys[i]].displayName < assets[keys[j]].displayName })
		if len(keys) > 0 {
			key = keys[0]
		}
	}
	asset, ok := assets[key]
	if !ok {
		return map[string]any{"mode": "media", "empty": true, "message": "Add media to preview it in context."}, nil
	}
	return map[string]any{
		"mode": "media", "media": previewMediaSummary(asset),
		"contexts": mediaPreviewContexts(definition, key),
	}, nil
}

func mediaPreviewContexts(definition DefinitionV1, key string) []map[string]any {
	contexts := []map[string]any{}
	if definition.Metadata.CoverAssetKey == key {
		contexts = append(contexts, map[string]any{
			"kind": "cover", "label": "Ruleset cover", "title": definition.Metadata.Name,
			"description": definition.Metadata.Description,
			"detail":      fmt.Sprintf("%d–%d players", definition.Metadata.MinPlayers, definition.Metadata.MaxPlayers),
		})
	}
	for _, team := range definition.Teams {
		if team.ImageAssetKey == key {
			contexts = append(contexts, map[string]any{"kind": "team", "label": "Team", "title": team.Name, "description": team.Description})
		}
	}
	for _, role := range definition.Roles {
		if role.ImageAssetKey != key {
			continue
		}
		teamName := "No team"
		for _, team := range definition.Teams {
			if team.ID == role.TeamID {
				teamName = team.Name
				break
			}
		}
		contexts = append(contexts, map[string]any{
			"kind": "role", "label": teamName, "title": role.Name,
			"description": role.Description, "detail": role.WinCondition,
		})
	}
	for _, ability := range definition.Abilities {
		if ability.ImageAssetKey == key {
			contexts = append(contexts, map[string]any{"kind": "ability", "label": "Ability", "title": ability.Name, "description": ability.Description})
		}
	}
	for _, achievement := range definition.Achievements {
		if achievement.ImageAssetKey == key {
			detail := fmt.Sprintf("%d achievement points", achievement.Points)
			if achievement.HiddenUntilGameCompleted {
				detail += " · Hidden until the game is complete"
			}
			contexts = append(contexts, map[string]any{"kind": "achievement", "label": "Achievement", "title": achievement.Name, "description": achievement.Description, "detail": detail})
		}
	}
	for _, cue := range definition.AudioCues {
		if cue.AssetKey != key {
			continue
		}
		used := false
		for _, phase := range definition.Phases {
			if phase.AudioCueID != cue.ID {
				continue
			}
			used = true
			detail := "Continues the current round"
			if phase.StartsRound {
				detail = "Starts a new round"
			}
			if phase.SuggestedDurationSeconds > 0 {
				detail += fmt.Sprintf(" · %d minute timer", int((time.Duration(phase.SuggestedDurationSeconds) * time.Second).Minutes()))
			}
			contexts = append(contexts, map[string]any{"kind": "phase", "label": "Game-master phase", "title": phase.Name, "description": phase.Description, "detail": detail})
		}
		if !used {
			contexts = append(contexts, map[string]any{"kind": "audio_cue", "label": "Audio cue", "title": cue.Name, "detail": "Available for announcements and future phase use"})
		}
	}
	return contexts
}

func previewMediaSummary(asset effectiveAsset) map[string]any {
	projected := projectEffectiveAsset(asset, nil)
	return map[string]any{
		"displayName": asset.displayName, "accessibilityText": asset.accessibilityText,
		"kind": asset.kind, "metadata": asset.metadata, "preview": projected["preview"],
	}
}

func selectedRole(definition DefinitionV1, id string) (Role, bool) {
	if id == "" && len(definition.Roles) > 0 {
		return definition.Roles[0], true
	}
	for _, role := range definition.Roles {
		if role.ID == id {
			return role, true
		}
	}
	return Role{}, false
}

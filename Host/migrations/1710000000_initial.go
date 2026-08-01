package migrations

import (
	"errors"

	"github.com/pocketbase/pocketbase/core"
	pbmigrations "github.com/pocketbase/pocketbase/migrations"
	"github.com/pocketbase/pocketbase/tools/types"
)

const (
	gameMastersCollection    = "game_masters"
	playerProfilesCollection = "player_profiles"
)

func init() {
	pbmigrations.Register(func(app core.App) error {
		gameMasters := core.NewAuthCollection(gameMastersCollection)
		lockRules(gameMasters)
		disableEmailRequirement(gameMasters)
		gameMasters.AuthRule = types.Pointer("active = true")
		gameMasters.PasswordAuth.Enabled = true
		gameMasters.PasswordAuth.IdentityFields = []string{"username"}
		gameMasters.AuthAlert.Enabled = false
		gameMasters.AuthToken.Duration = 12 * 60 * 60
		if password, ok := gameMasters.Fields.GetByName(core.FieldNamePassword).(*core.PasswordField); ok {
			password.Min = 6
		}
		gameMasters.Fields.Add(
			text("username", true, 3, 32),
			text("display_name", true, 2, 64),
			&core.BoolField{Name: "is_owner"},
			&core.BoolField{Name: "active"},
			&core.DateField{Name: "last_login_at"},
		)
		gameMasters.AddIndex("idx_game_masters_username", true, "username", "")
		if err := app.Save(gameMasters); err != nil {
			return err
		}

		hostSettings := core.NewBaseCollection("host_settings")
		lockRules(hostSettings)
		hostSettings.Fields.Add(
			&core.NumberField{Name: "port", Required: true, OnlyInt: true, Min: number(1), Max: number(65535)},
			&core.TextField{Name: "bind_address", Max: 45},
			&core.TextField{Name: "preferred_adapter", Max: 200},
			&core.BoolField{Name: "trusted_lan_acknowledged"},
			&core.BoolField{Name: "automatic_backups"},
		)
		if err := app.Save(hostSettings); err != nil {
			return err
		}

		playerProfiles := core.NewAuthCollection(playerProfilesCollection)
		lockRules(playerProfiles)
		disableEmailRequirement(playerProfiles)
		playerProfiles.AuthRule = types.Pointer("active = true")
		playerProfiles.PasswordAuth.Enabled = false
		playerProfiles.AuthAlert.Enabled = false
		playerProfiles.AuthToken.Duration = 180 * 24 * 60 * 60
		playerProfiles.Fields.Add(
			text("display_name", true, 2, 32),
			&core.TextField{Name: "normalized_name", Required: true, Hidden: true, Max: 64},
			&core.FileField{Name: "avatar", MaxSize: 1 << 20, MaxSelect: 1, MimeTypes: []string{"image/jpeg", "image/png", "image/webp"}, Protected: true},
			&core.TextField{Name: "bio", Max: 280},
			&core.SelectField{Name: "accent", Values: []string{"crimson", "forest", "navy", "gold", "plum"}},
			&core.BoolField{Name: "active"},
			&core.DateField{Name: "approved_at"},
			relation("approved_by", gameMasters.Id, false),
			&core.DateField{Name: "last_seen_at"},
		)
		playerProfiles.AddIndex("idx_player_profiles_normalized_name", true, "normalized_name", "")
		if err := app.Save(playerProfiles); err != nil {
			return err
		}

		profileRequests := core.NewBaseCollection("profile_requests")
		lockRules(profileRequests)
		profileRequests.Fields.Add(
			selectField("request_type", true, "new", "recover"),
			text("requested_name", true, 2, 32),
			&core.TextField{Name: "normalized_name", Required: true, Hidden: true, Max: 64},
			relation("profile", playerProfiles.Id, false),
			&core.TextField{Name: "secret_hash", Required: true, Hidden: true, Max: 255},
			selectField("status", true, "pending", "approved", "rejected", "expired", "consumed"),
			&core.DateField{Name: "expires_at", Required: true},
			relation("decided_by", gameMasters.Id, false),
			&core.DateField{Name: "decided_at"},
			&core.DateField{Name: "consumed_at"},
			&core.TextField{Name: "rejection_reason", Max: 280},
		)
		profileRequests.AddIndex("idx_profile_requests_expiry", false, "status,expires_at", "")
		if err := app.Save(profileRequests); err != nil {
			return err
		}

		rulesets := core.NewBaseCollection("rulesets")
		lockRules(rulesets)
		rulesets.Fields.Add(
			text("slug", true, 1, 64),
			text("name", true, 1, 120),
			&core.BoolField{Name: "archived"},
			relation("created_by", gameMasters.Id, true),
		)
		rulesets.AddIndex("idx_rulesets_slug", true, "slug", "")
		if err := app.Save(rulesets); err != nil {
			return err
		}

		rulesetVersions := core.NewBaseCollection("ruleset_versions")
		lockRules(rulesetVersions)
		rulesetVersions.Fields.Add(
			relation("ruleset", rulesets.Id, true),
			&core.NumberField{Name: "version_number", Required: true, OnlyInt: true, Min: number(1)},
			selectField("state", true, "draft", "published"),
			&core.NumberField{Name: "schema_version", Required: true, OnlyInt: true, Min: number(1)},
			&core.JSONField{Name: "definition", Required: true, MaxSize: 2 << 20},
			&core.TextField{Name: "definition_checksum", Max: 64},
			relation("created_by", gameMasters.Id, true),
			relation("published_by", gameMasters.Id, false),
			&core.DateField{Name: "published_at"},
			&core.JSONField{Name: "source_metadata", MaxSize: 64 << 10},
		)
		rulesetVersions.AddIndex("idx_ruleset_versions_number", true, "ruleset,version_number", "")
		rulesetVersions.AddIndex("idx_ruleset_versions_one_draft", true, "ruleset", "state = 'draft'")
		if err := app.Save(rulesetVersions); err != nil {
			return err
		}

		rulesets.Fields.Add(relation("latest_published_version", rulesetVersions.Id, false))
		if err := app.Save(rulesets); err != nil {
			return err
		}

		rulesetAssets := core.NewBaseCollection("ruleset_assets")
		lockRules(rulesetAssets)
		rulesetAssets.Fields.Add(
			relation("ruleset_version", rulesetVersions.Id, true),
			text("asset_key", true, 1, 64),
			selectField("kind", true, "image", "audio"),
			&core.FileField{Name: "file", MaxSize: 5 << 20, MaxSelect: 1, MimeTypes: []string{"image/jpeg", "image/png", "image/webp", "audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav"}, Protected: true},
			selectField("storage_state", true, "staging", "ready"),
			text("mime_type", true, 1, 100),
			&core.TextField{Name: "checksum", Required: true, Max: 64},
			&core.JSONField{Name: "metadata", MaxSize: 16 << 10},
		)
		rulesetAssets.AddIndex("idx_ruleset_assets_key", true, "ruleset_version,asset_key", "")
		if err := app.Save(rulesetAssets); err != nil {
			return err
		}

		games := core.NewBaseCollection("games")
		lockRules(games)
		games.Fields.Add(
			text("name", true, 1, 120),
			selectField("status", true, "draft", "lobby", "running", "paused", "review", "archived"),
			relation("ruleset_version", rulesetVersions.Id, true),
			&core.JSONField{Name: "ruleset_snapshot", Required: true, MaxSize: 2 << 20},
			&core.TextField{Name: "join_code", Max: 6},
			&core.BoolField{Name: "joining_open"},
			intField("revision", 0),
			intField("round_number", 0),
			&core.TextField{Name: "phase_key", Max: 32},
			&core.DateField{Name: "phase_started_at"},
			selectField("timer_state", true, "inactive", "running", "paused", "completed"),
			intField("timer_total_ms", 0),
			intField("timer_remaining_ms", 0),
			&core.DateField{Name: "timer_ends_at"},
			intField("timer_revision", 0),
			&core.DateField{Name: "started_at"},
			&core.DateField{Name: "ended_at"},
			relation("created_by", gameMasters.Id, true),
		)
		games.AddIndex("idx_games_join_code", true, "join_code", "join_code != ''")
		if err := app.Save(games); err != nil {
			return err
		}
		if _, err := app.DB().NewQuery(`
			CREATE UNIQUE INDEX idx_games_single_live
			ON games ((1))
			WHERE status IN ('lobby','running','paused')
		`).Execute(); err != nil {
			return err
		}

		participants := core.NewBaseCollection("participants")
		lockRules(participants)
		participants.Fields.Add(
			relation("game", games.Id, true),
			relation("profile", playerProfiles.Id, true),
			text("display_name_snapshot", true, 2, 32),
			&core.TextField{Name: "game_alias", Max: 32},
			&core.NumberField{Name: "seat_number", Required: true, OnlyInt: true, Min: number(1), Max: number(30)},
			selectField("status", true, "active", "eliminated", "kicked", "left"),
			&core.TextField{Name: "role_key", Hidden: true, Max: 32},
			selectField("outcome", true, "unset", "win", "loss", "draw"),
			&core.DateField{Name: "joined_at", Required: true},
			&core.DateField{Name: "eliminated_at"},
			relation("assigned_by", gameMasters.Id, false),
		)
		participants.AddIndex("idx_participants_game_profile", true, "game,profile", "")
		participants.AddIndex("idx_participants_game_seat", true, "game,seat_number", "")
		if err := app.Save(participants); err != nil {
			return err
		}

		chatRooms := core.NewBaseCollection("chat_rooms")
		lockRules(chatRooms)
		chatRooms.Fields.Add(
			relation("game", games.Id, true),
			text("room_key", true, 1, 120),
			selectField("kind", true, "announcements", "gm_dm", "general", "team", "player_dm"),
			text("label", true, 1, 120),
			&core.TextField{Name: "team_key", Max: 32},
			&core.BoolField{Name: "manually_locked"},
			selectField("manual_visibility_override", true, "default", "visible", "hidden"),
			selectField("sender_display", true, "profile_name", "game_alias", "seat_number", "role_label", "team_label"),
		)
		chatRooms.AddIndex("idx_chat_rooms_key", true, "game,room_key", "")
		if err := app.Save(chatRooms); err != nil {
			return err
		}

		chatMemberships := core.NewBaseCollection("chat_memberships")
		lockRules(chatMemberships)
		chatMemberships.Fields.Add(
			relation("room", chatRooms.Id, true),
			relation("participant", participants.Id, true),
			&core.DateField{Name: "joined_at", Required: true},
			&core.DateField{Name: "left_at"},
			&core.BoolField{Name: "historical_access"},
		)
		chatMemberships.AddIndex("idx_chat_memberships_unique", true, "room,participant", "")
		if err := app.Save(chatMemberships); err != nil {
			return err
		}

		chatMessages := core.NewBaseCollection("chat_messages")
		lockRules(chatMessages)
		chatMessages.Fields.Add(
			relation("room", chatRooms.Id, true),
			selectField("message_kind", true, "message", "announcement", "system"),
			selectField("sender_type", true, "player", "game_master", "system"),
			&core.TextField{Name: "sender_id", Hidden: true, Max: 64},
			relation("sender_participant", participants.Id, false),
			text("sender_label_snapshot", true, 1, 120),
			&core.TextField{Name: "content", Max: 1000},
			&core.TextField{Name: "cue_key", Max: 32},
			&core.DateField{Name: "deleted_at"},
			relation("deleted_by", gameMasters.Id, false),
		)
		chatMessages.AddIndex("idx_chat_messages_cursor", false, "room,created DESC,id DESC", "")
		chatMessages.AddIndex("idx_chat_messages_sender", false, "sender_participant,created DESC", "")
		if err := app.Save(chatMessages); err != nil {
			return err
		}

		achievementAwards := core.NewBaseCollection("achievement_awards")
		lockRules(achievementAwards)
		achievementAwards.Fields.Add(
			relation("profile", playerProfiles.Id, true),
			relation("game", games.Id, true),
			relation("ruleset_version", rulesetVersions.Id, true),
			text("achievement_key", true, 1, 32),
			text("title_snapshot", true, 1, 120),
			&core.TextField{Name: "description_snapshot", Max: 1000},
			&core.TextField{Name: "asset_key", Max: 64},
			&core.NumberField{Name: "points_snapshot", OnlyInt: true, Min: number(0), Max: number(10000)},
			&core.BoolField{Name: "hidden_until_game_completed"},
			relation("awarded_by", gameMasters.Id, true),
			&core.TextField{Name: "note", Hidden: true, Max: 1000},
		)
		achievementAwards.AddIndex("idx_achievement_awards_unique", true, "profile,game,achievement_key", "")
		if err := app.Save(achievementAwards); err != nil {
			return err
		}

		gameAudit := core.NewBaseCollection("game_audit")
		lockRules(gameAudit)
		gameAudit.Fields.Add(
			relation("game", games.Id, false),
			selectField("actor_type", true, "game_master", "player", "system"),
			&core.TextField{Name: "actor_id", Hidden: true, Max: 64},
			&core.TextField{Name: "actor_label", Max: 120},
			text("action", true, 1, 120),
			&core.TextField{Name: "target_type", Max: 64},
			&core.TextField{Name: "target_id", Hidden: true, Max: 64},
			&core.JSONField{Name: "detail", MaxSize: 64 << 10},
			&core.TextField{Name: "request_id", Max: 64},
		)
		gameAudit.AddIndex("idx_game_audit_game_created", false, "game,created DESC", "")
		return app.Save(gameAudit)
	}, func(app core.App) error {
		rulesets, err := app.FindCollectionByNameOrId("rulesets")
		if err == nil {
			rulesets.Fields.RemoveByName("latest_published_version")
			if err := app.Save(rulesets); err != nil {
				return err
			}
		}
		names := []string{
			"game_audit",
			"achievement_awards",
			"chat_messages",
			"chat_memberships",
			"chat_rooms",
			"participants",
			"games",
			"ruleset_assets",
			"ruleset_versions",
			"rulesets",
			"profile_requests",
			playerProfilesCollection,
			"host_settings",
			gameMastersCollection,
		}
		for _, name := range names {
			collection, err := app.FindCollectionByNameOrId(name)
			if err != nil {
				continue
			}
			if err := app.Delete(collection); err != nil {
				return errors.New("failed to remove collection " + name + ": " + err.Error())
			}
		}
		return nil
	}, "1710000000_initial.go")
}

func lockRules(collection *core.Collection) {
	collection.ListRule = nil
	collection.ViewRule = nil
	collection.CreateRule = nil
	collection.UpdateRule = nil
	collection.DeleteRule = nil
	collection.Fields.Add(
		&core.AutodateField{Name: "created", OnCreate: true},
		&core.AutodateField{Name: "updated", OnCreate: true, OnUpdate: true},
	)
}

func text(name string, required bool, min, max int) *core.TextField {
	return &core.TextField{Name: name, Required: required, Min: min, Max: max}
}

func relation(name, collectionID string, required bool) *core.RelationField {
	field := &core.RelationField{Name: name, CollectionId: collectionID, MaxSelect: 1, Required: required}
	if required {
		field.MinSelect = 1
	}
	return field
}

func selectField(name string, required bool, values ...string) *core.SelectField {
	return &core.SelectField{Name: name, Required: required, Values: values, MaxSelect: 1}
}

func intField(name string, min float64) *core.NumberField {
	return &core.NumberField{Name: name, OnlyInt: true, Min: &min}
}

func number(value float64) *float64 {
	return &value
}

func disableEmailRequirement(collection *core.Collection) {
	if field, ok := collection.Fields.GetByName(core.FieldNameEmail).(*core.EmailField); ok {
		field.Required = false
		field.Hidden = true
	}
}

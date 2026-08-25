package rulesets

type DefinitionV1 struct {
	SchemaVersion        int                           `json:"schemaVersion"`
	Metadata             Metadata                      `json:"metadata"`
	Teams                []Team                        `json:"teams"`
	Categories           []Category                    `json:"categories"`
	Abilities            []Ability                     `json:"abilities"`
	Roles                []Role                        `json:"roles"`
	Phases               []Phase                       `json:"phases"`
	KnowledgeRules       []KnowledgeRule               `json:"knowledgeRules"`
	CompositionBands     []CompositionBand             `json:"compositionBands"`
	CompositionModifiers []CompositionModifier         `json:"compositionModifiers"`
	Chat                 ChatPolicy                    `json:"chat"`
	Achievements         []Achievement                 `json:"achievements"`
	AudioCues            []AudioCue                    `json:"audioCues"`
	AssetAccessibility   map[string]AssetAccessibility `json:"assetAccessibility,omitempty"`
}

type Metadata struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	MinPlayers    int    `json:"minPlayers"`
	MaxPlayers    int    `json:"maxPlayers"`
	CoverAssetKey string `json:"coverAssetKey,omitempty"`
}

type Team struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	ImageAssetKey string `json:"imageAssetKey,omitempty"`
}

type Category struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

type Ability struct {
	ID                           string   `json:"id"`
	Name                         string   `json:"name"`
	Description                  string   `json:"description"`
	ImageAssetKey                string   `json:"imageAssetKey,omitempty"`
	ActivationPhaseIDs           []string `json:"activationPhaseIds,omitempty"`
	CanCombineWithOtherAbilities bool     `json:"canCombineWithOtherAbilities"`
}

type Role struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	TeamID        string   `json:"teamId"`
	CategoryIDs   []string `json:"categoryIds"`
	Tags          []string `json:"tags"`
	AbilityIDs    []string `json:"abilityIds"`
	WinCondition  string   `json:"winCondition"`
	MaxCopies     int      `json:"maxCopies"`
	ImageAssetKey string   `json:"imageAssetKey,omitempty"`
}

type Phase struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	Order                    int    `json:"order"`
	StartsRound              bool   `json:"startsRound"`
	SuggestedDurationSeconds int    `json:"suggestedDurationSeconds,omitempty"`
	AudioCueID               string `json:"audioCueId,omitempty"`
}

type Selector struct {
	RoleIDs     []string `json:"roleIds,omitempty"`
	TeamIDs     []string `json:"teamIds,omitempty"`
	CategoryIDs []string `json:"categoryIds,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type KnowledgeRule struct {
	Viewer Selector `json:"viewer"`
	Target Selector `json:"target"`
	Reveal []string `json:"reveal"`
}

type CompositionBand struct {
	ID         string            `json:"id"`
	MinPlayers int               `json:"minPlayers"`
	MaxPlayers int               `json:"maxPlayers"`
	Slots      []CompositionSlot `json:"slots"`
}

type CompositionSlot struct {
	ID       string   `json:"id"`
	Label    string   `json:"label"`
	Count    int      `json:"count"`
	Selector Selector `json:"selector"`
}

type CompositionModifier struct {
	ID              string           `json:"id"`
	WhenRolePresent string           `json:"whenRolePresent"`
	SlotAdjustments []SlotAdjustment `json:"slotAdjustments"`
	RequiresRoleIDs []string         `json:"requiresRoleIds"`
	ExcludesRoleIDs []string         `json:"excludesRoleIds"`
}

type SlotAdjustment struct {
	SlotID string `json:"slotId"`
	Delta  int    `json:"delta"`
}

type SenderDisplay string

const (
	SenderProfileName SenderDisplay = "profile_name"
	SenderGameAlias   SenderDisplay = "game_alias"
	SenderSeatNumber  SenderDisplay = "seat_number"
	SenderRoleLabel   SenderDisplay = "role_label"
	SenderTeamLabel   SenderDisplay = "team_label"
)

type RoomPermission struct {
	Visible           bool          `json:"visible"`
	Readable          bool          `json:"readable"`
	Sendable          bool          `json:"sendable"`
	GameMasterMaySend bool          `json:"gameMasterMaySend"`
	SenderDisplay     SenderDisplay `json:"senderDisplay"`
}

type PartialRoomPermission struct {
	Visible           *bool          `json:"visible,omitempty"`
	Readable          *bool          `json:"readable,omitempty"`
	Sendable          *bool          `json:"sendable,omitempty"`
	GameMasterMaySend *bool          `json:"gameMasterMaySend,omitempty"`
	SenderDisplay     *SenderDisplay `json:"senderDisplay,omitempty"`
}

type ChatPolicy struct {
	DefaultPolicy  ChatPolicyDefaults            `json:"defaultPolicy"`
	PhaseOverrides map[string]ChatPolicyOverride `json:"phaseOverrides"`
	Channels       []ChatChannel                 `json:"channels,omitempty"`
}

type ChatPolicyDefaults struct {
	General  *RoomPermission           `json:"general,omitempty"`
	PlayerDM *RoomPermission           `json:"playerDm,omitempty"`
	Teams    map[string]RoomPermission `json:"teams"`
}

type ChatPolicyOverride struct {
	General  *PartialRoomPermission           `json:"general,omitempty"`
	PlayerDM *PartialRoomPermission           `json:"playerDm,omitempty"`
	Teams    map[string]PartialRoomPermission `json:"teams,omitempty"`
}

type ChatMessageRestriction string

const (
	ChatNormalText ChatMessageRestriction = "normal_text"
	ChatEmojiOnly  ChatMessageRestriction = "emoji_only"
)

// ChatChannel is an additional ruleset-defined room. An empty reader selector
// means every assigned role and an empty sender selector means every reader;
// otherwise matching any listed role or team grants the permission.
type ChatChannel struct {
	ID                 string                              `json:"id"`
	Name               string                              `json:"name"`
	ReaderRoleIDs      []string                            `json:"readerRoleIds"`
	ReaderTeamIDs      []string                            `json:"readerTeamIds"`
	SenderRoleIDs      []string                            `json:"senderRoleIds"`
	SenderTeamIDs      []string                            `json:"senderTeamIds"`
	MessageRestriction ChatMessageRestriction              `json:"messageRestriction"`
	Visible            bool                                `json:"visible"`
	Sendable           bool                                `json:"sendable"`
	GameMasterMaySend  bool                                `json:"gameMasterMaySend"`
	SenderDisplay      SenderDisplay                       `json:"senderDisplay"`
	PhaseOverrides     map[string]ChatChannelPhaseOverride `json:"phaseOverrides"`
}

type ChatChannelPhaseOverride struct {
	Visible  *bool `json:"visible,omitempty"`
	Sendable *bool `json:"sendable,omitempty"`
}

type Achievement struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	ImageAssetKey            string `json:"imageAssetKey,omitempty"`
	Points                   int    `json:"points"`
	HiddenUntilGameCompleted bool   `json:"hiddenUntilGameCompleted"`
}

type AudioCue struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	AssetKey        string `json:"assetKey"`
	DefaultAudience string `json:"defaultAudience"`
}

type AssetAccessibility struct {
	Description string `json:"description"`
}

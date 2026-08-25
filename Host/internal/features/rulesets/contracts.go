package rulesets

import domainrulesets "github.com/jvanspijk/SocialGamesHoster/Host/internal/domain/rulesets"

type (
	DefinitionV1             = domainrulesets.DefinitionV1
	Metadata                 = domainrulesets.Metadata
	Team                     = domainrulesets.Team
	Category                 = domainrulesets.Category
	Ability                  = domainrulesets.Ability
	Role                     = domainrulesets.Role
	Phase                    = domainrulesets.Phase
	Selector                 = domainrulesets.Selector
	KnowledgeRule            = domainrulesets.KnowledgeRule
	CompositionBand          = domainrulesets.CompositionBand
	CompositionSlot          = domainrulesets.CompositionSlot
	CompositionModifier      = domainrulesets.CompositionModifier
	SlotAdjustment           = domainrulesets.SlotAdjustment
	SenderDisplay            = domainrulesets.SenderDisplay
	RoomPermission           = domainrulesets.RoomPermission
	PartialRoomPermission    = domainrulesets.PartialRoomPermission
	ChatPolicy               = domainrulesets.ChatPolicy
	ChatPolicyDefaults       = domainrulesets.ChatPolicyDefaults
	ChatPolicyOverride       = domainrulesets.ChatPolicyOverride
	ChatMessageRestriction   = domainrulesets.ChatMessageRestriction
	ChatChannel              = domainrulesets.ChatChannel
	ChatChannelPhaseOverride = domainrulesets.ChatChannelPhaseOverride
	Achievement              = domainrulesets.Achievement
	AudioCue                 = domainrulesets.AudioCue
	AssetAccessibility       = domainrulesets.AssetAccessibility
)

const (
	SenderProfileName = domainrulesets.SenderProfileName
	SenderGameAlias   = domainrulesets.SenderGameAlias
	SenderSeatNumber  = domainrulesets.SenderSeatNumber
	SenderRoleLabel   = domainrulesets.SenderRoleLabel
	SenderTeamLabel   = domainrulesets.SenderTeamLabel
	ChatNormalText    = domainrulesets.ChatNormalText
	ChatEmojiOnly     = domainrulesets.ChatEmojiOnly
)

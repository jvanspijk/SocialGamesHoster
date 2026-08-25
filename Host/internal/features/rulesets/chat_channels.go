package rulesets

import "strings"

const CustomChatRoomPrefix = "custom:"

func FindChatChannel(definition DefinitionV1, id string) *ChatChannel {
	for index := range definition.Chat.Channels {
		if definition.Chat.Channels[index].ID == id {
			return &definition.Chat.Channels[index]
		}
	}
	return nil
}

func ChatChannelIDFromRoomKey(roomKey string) string {
	return strings.TrimPrefix(roomKey, CustomChatRoomPrefix)
}

func ChatChannelAudienceMatches(channel ChatChannel, role Role, sender bool) bool {
	roleIDs := channel.ReaderRoleIDs
	teamIDs := channel.ReaderTeamIDs
	if sender {
		roleIDs = channel.SenderRoleIDs
		teamIDs = channel.SenderTeamIDs
		if len(roleIDs) == 0 && len(teamIDs) == 0 {
			return ChatChannelAudienceMatches(channel, role, false)
		}
	}
	if len(roleIDs) == 0 && len(teamIDs) == 0 {
		return true
	}
	for _, id := range roleIDs {
		if id == role.ID {
			return true
		}
	}
	for _, id := range teamIDs {
		if id == role.TeamID {
			return true
		}
	}
	return false
}

func ChatChannelPolicy(channel ChatChannel, phaseID string) (RoomPermission, *PartialRoomPermission) {
	base := RoomPermission{
		Visible: channel.Visible, Readable: channel.Visible, Sendable: channel.Sendable,
		GameMasterMaySend: channel.GameMasterMaySend, SenderDisplay: channel.SenderDisplay,
	}
	phase, ok := channel.PhaseOverrides[phaseID]
	if !ok {
		return base, nil
	}
	override := PartialRoomPermission{Visible: phase.Visible, Readable: phase.Visible, Sendable: phase.Sendable}
	return base, &override
}

// ApplyRoomOverride is shared by live chat and the authoring preview so both
// resolve phase-specific permissions identically.
func ApplyRoomOverride(base RoomPermission, override *PartialRoomPermission) RoomPermission {
	if override == nil {
		return base
	}
	if override.Visible != nil {
		base.Visible = *override.Visible
	}
	if override.Readable != nil {
		base.Readable = *override.Readable
	}
	if override.Sendable != nil {
		base.Sendable = *override.Sendable
	}
	if override.GameMasterMaySend != nil {
		base.GameMasterMaySend = *override.GameMasterMaySend
	}
	if override.SenderDisplay != nil {
		base.SenderDisplay = *override.SenderDisplay
	}
	return base
}

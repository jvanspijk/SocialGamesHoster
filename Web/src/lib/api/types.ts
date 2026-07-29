export type ActorType = 'game_masters' | 'player_profiles';

export interface Actor {
	id: string;
	type: ActorType;
	displayName: string;
	isOwner?: boolean;
}

export interface AuthResponse {
	token: string;
	actor?: Actor;
	profile?: Profile;
}

export interface Profile {
	id: string;
	displayName: string;
	avatar: string;
	bio: string;
	accent: string;
	active: boolean;
}

export interface ProfileRequest {
	id: string;
	requestType: 'new' | 'recover';
	requestedName: string;
	createdAt: string;
	expiresAt: string;
}

export interface AppErrorBody {
	code: string;
	message: string;
	fieldErrors?: Record<string, string[]>;
	traceId?: string;
}

export interface Game {
	id: string;
	name: string;
	status: 'draft' | 'lobby' | 'running' | 'paused' | 'review' | 'archived';
	rulesetVersion: string;
	/** @deprecated Join codes are no longer part of reader projections. */
	joinCode?: string;
	joiningOpen: boolean;
	rolesVisible: boolean;
	roleVisibilityRevision: number;
	revision: number;
	roundNumber: number;
	phaseKey: string;
	phaseStartedAt?: string;
	abilityPhaseLockedAt?: string;
	startedAt?: string;
	endedAt?: string;
}

export interface Participant {
	id: string;
	profileId: string;
	displayNameSnapshot: string;
	gameAlias: string;
	seatNumber: number;
	status: 'active' | 'eliminated' | 'kicked' | 'left';
	outcome: 'unset' | 'win' | 'loss' | 'draw';
	roleKey?: string;
	roleRevision?: number;
}

export interface Room {
	id: string;
	key: string;
	kind: string;
	label: string;
	playersCanPost: boolean;
	/** @deprecated Use playersCanPost. */
	locked?: boolean;
	readable: boolean;
	sendable: boolean;
	gameMasterMaySend?: boolean;
	messageRestriction?: 'normal_text' | 'emoji_only';
	latestMessage: MessageSummary | null;
}

export interface MessageSummary {
	createdAt: string;
	id: string;
	senderLabel: string;
	preview: string;
}

export interface ChatMessage {
	id: string;
	roomId: string;
	kind: string;
	senderType: string;
	senderLabel: string;
	content: string;
	cueKey?: string;
	deleted: boolean;
	createdAt: string;
	isOwn?: boolean;
}

export interface AnnouncementAttentionItem {
	id: string;
	kind: 'announcement';
	senderLabel: string;
	content: string;
	cueKey?: string;
	createdAt: string;
	image?: { url: string; description: string };
	audio?: { url: string; alternative: string };
}

export interface FutureEventAttentionItem {
	id: string;
	kind: 'event';
	createdAt: string;
}

export type AttentionItem = AnnouncementAttentionItem | FutureEventAttentionItem;

export interface AdminAttentionSummary extends AnnouncementAttentionItem {
	audience: 'all' | 'team' | 'player';
	targetId?: string;
	recipientTotal: number;
	acknowledgementCount: number;
}

export interface TimerProjection {
	status: 'inactive' | 'running' | 'paused' | 'completed';
	totalMs: number;
	remainingMs: number;
	endsAt?: string;
	revision: number;
	serverTime: string;
}

export interface RoleProjection {
	id: string;
	name: string;
	description: string;
	winCondition: string;
	imageAssetKey?: string;
	team?: { id: string; name: string; description: string };
	abilities: RulesetAbility[];
}

export interface AbilityChoice {
	id: string;
	abilityId: string;
	abilityName: string;
	status: 'Activated' | 'Finalized';
	activatedAt: string;
	finalizedAt?: string;
}

export interface PlayerGameView {
	game: Game;
	participant: {
		id: string;
		displayName: string;
		gameAlias: string;
		seatNumber: number;
		status: string;
	};
	ruleset: { name: string; description: string };
	roleAvailable: boolean;
	roleRevision: number;
	role: RoleProjection | null;
	knowledge: Array<Record<string, unknown>>;
	rooms: Room[];
	attentionItems: AttentionItem[];
	assets: Array<{
		id: string;
		assetKey: string;
		kind: 'image' | 'audio';
		checksum: string;
		preview: string;
	}>;
	party: Array<{
		id: string;
		profileId: string;
		displayName: string;
		gameAlias: string;
		seatNumber: number;
		status: string;
	}>;
	abilityChoices?: AbilityChoice[];
}

export interface AdminGameView {
	game: Game;
	timer: TimerProjection;
	ruleset: RulesetDefinition;
	participants: Participant[];
	rooms: Room[];
	attentionItems: AdminAttentionSummary[];
	awards: Array<{
		id: string;
		profileId: string;
		achievementId: string;
		title: string;
		description: string;
		points: number;
		hiddenUntilGameCompleted: boolean;
		awardedAt: string;
	}>;
	audit: Array<{
		id: string;
		actorLabel: string;
		action: string;
		targetType: string;
		detail?: Record<string, unknown>;
		createdAt: string;
	}>;
	assets: Array<{ id: string; assetKey: string; kind: 'image' | 'audio' }>;
	abilityProgress: {
		phaseKey: string;
		roundNumber: number;
		locked: boolean;
		lockedAt?: string;
		eligiblePlayerCount: number;
		activatedPlayerCount: number;
		finalizedPlayerCount: number;
	};
	abilityResults: Array<{
		participantId: string;
		displayName: string;
		seatNumber: number;
		phaseKey: string;
		roundNumber: number;
		abilities: Array<{ id: string; name: string }>;
	}>;
}

export interface ActivityItem {
	id: string;
	text: string;
	createdAt: string;
}

export interface GameSummary {
	game: Game;
	ruleset: { name: string; coverAssetKey?: string };
	durationMs: number;
	participants: Array<
		Participant & {
			achievements: Array<{ id: string; title: string; description: string; points: number }>;
		}
	>;
	immutable: boolean;
}

export interface RulesetSummary {
	id: string;
	slug: string;
	name: string;
	archived: boolean;
	latestPublishedVersion: string;
}

export interface RulesetDefinition {
	schemaVersion: 1;
	metadata: {
		name: string;
		description: string;
		minPlayers: number;
		maxPlayers: number;
		coverAssetKey?: string;
	};
	teams: RulesetTeam[];
	categories: RulesetCategory[];
	abilities: RulesetAbility[];
	roles: RulesetRole[];
	phases: RulesetPhase[];
	knowledgeRules: RulesetKnowledgeRule[];
	compositionBands: RulesetCompositionBand[];
	compositionModifiers: RulesetCompositionModifier[];
	chat: RulesetChatPolicy;
	achievements: RulesetAchievement[];
	audioCues: RulesetAudioCue[];
	assetAccessibility?: Record<string, { description: string }>;
}

export interface RulesetTeam {
	id: string;
	name: string;
	description: string;
	imageAssetKey?: string;
}

export interface RulesetCategory {
	id: string;
	name: string;
	description?: string;
}

export interface RulesetAbility {
	id: string;
	name: string;
	description: string;
	imageAssetKey?: string;
	activationPhaseIds?: string[];
	canCombineWithOtherAbilities?: boolean;
}

export interface RulesetRole {
	id: string;
	name: string;
	description: string;
	teamId: string;
	categoryIds: string[];
	tags: string[];
	abilityIds: string[];
	winCondition: string;
	maxCopies: number;
	imageAssetKey?: string;
}

export interface RulesetPhase {
	id: string;
	name: string;
	description: string;
	order: number;
	startsRound: boolean;
	suggestedDurationSeconds?: number;
	audioCueId?: string;
}

export interface RulesetSelector {
	roleIds: string[];
	teamIds: string[];
	categoryIds: string[];
	tags: string[];
}

export interface RulesetKnowledgeRule {
	viewer: RulesetSelector;
	target: RulesetSelector;
	reveal: string[];
}

export interface RulesetCompositionBand {
	id: string;
	minPlayers: number;
	maxPlayers: number;
	slots: Array<{
		id: string;
		label: string;
		count: number;
		selector: RulesetSelector;
	}>;
}

export interface RulesetCompositionModifier {
	id: string;
	whenRolePresent: string;
	slotAdjustments: Array<{ slotId: string; delta: number }>;
	requiresRoleIds: string[];
	excludesRoleIds: string[];
}

export type RulesetSenderDisplay =
	| 'profile_name'
	| 'game_alias'
	| 'seat_number'
	| 'role_label'
	| 'team_label';

export interface RulesetRoomPermission {
	visible: boolean;
	readable: boolean;
	sendable: boolean;
	gameMasterMaySend: boolean;
	senderDisplay: RulesetSenderDisplay;
}

export interface RulesetPartialRoomPermission {
	visible?: boolean;
	readable?: boolean;
	sendable?: boolean;
	gameMasterMaySend?: boolean;
	senderDisplay?: RulesetSenderDisplay;
}

export interface RulesetChatPolicy {
	defaultPolicy: {
		general?: RulesetRoomPermission;
		playerDm?: RulesetRoomPermission;
		teams: Record<string, RulesetRoomPermission>;
	};
	phaseOverrides: Record<
		string,
		{
			general?: RulesetPartialRoomPermission;
			playerDm?: RulesetPartialRoomPermission;
			teams?: Record<string, RulesetPartialRoomPermission>;
		}
	>;
	channels: RulesetChatChannel[];
}

export interface RulesetChatChannel {
	id: string;
	name: string;
	readerRoleIds: string[];
	readerTeamIds: string[];
	senderRoleIds: string[];
	senderTeamIds: string[];
	messageRestriction: 'normal_text' | 'emoji_only';
	visible: boolean;
	sendable: boolean;
	gameMasterMaySend: boolean;
	senderDisplay: RulesetSenderDisplay;
	phaseOverrides: Record<
		string,
		{
			visible?: boolean;
			sendable?: boolean;
		}
	>;
}

export interface RulesetAchievement {
	id: string;
	name: string;
	description: string;
	imageAssetKey?: string;
	points: number;
	hiddenUntilGameCompleted: boolean;
}

export interface RulesetAudioCue {
	id: string;
	name: string;
	assetKey: string;
	defaultAudience: 'all' | 'team' | 'player' | 'game_masters';
}

export interface RealtimeEnvelope<T = unknown> {
	eventId: string;
	gameId?: string;
	revision?: number;
	kind: string;
	occurredAt: string;
	payload: T;
}

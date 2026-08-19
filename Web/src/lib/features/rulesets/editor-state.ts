import type { RulesetDefinition } from '$lib/api/types';

export type EditorSection =
	| 'metadata'
	| 'teams'
	| 'roles'
	| 'composition'
	| 'phases'
	| 'knowledge'
	| 'chat'
	| 'achievements'
	| 'audio';

export type ValidationIssue = { path: string; code?: string; message: string };
export type ValidationReport = { errors: ValidationIssue[]; warnings: ValidationIssue[] };
export type SectionState = 'Not started' | 'In progress' | 'Complete' | 'Needs attention';
export type RecoveryRecord = {
	version: 1;
	definition: RulesetDefinition;
	section: EditorSection;
	selectedItems: Record<string, string>;
	timestamp: string;
};

export const requiredSections: EditorSection[] = ['metadata', 'teams', 'roles', 'composition'];

export function copyDefinition(definition: RulesetDefinition): RulesetDefinition {
	const value = JSON.parse(JSON.stringify(definition)) as RulesetDefinition;
	return {
		...value,
		teams: value.teams ?? [],
		categories: value.categories ?? [],
		abilities: value.abilities ?? [],
		roles: value.roles ?? [],
		phases: value.phases ?? [],
		knowledgeRules: value.knowledgeRules ?? [],
		compositionBands: value.compositionBands ?? [],
		compositionModifiers: value.compositionModifiers ?? [],
		chat: {
			defaultPolicy: {
				...(value.chat?.defaultPolicy ?? {}),
				teams: value.chat?.defaultPolicy?.teams ?? {}
			},
			phaseOverrides: value.chat?.phaseOverrides ?? {},
			channels: value.chat?.channels ?? []
		},
		achievements: value.achievements ?? [],
		audioCues: value.audioCues ?? [],
		assetAccessibility: value.assetAccessibility ?? {}
	};
}

export function normalizeReport(
	report: Partial<ValidationReport> | null | undefined
): ValidationReport {
	return { errors: report?.errors ?? [], warnings: report?.warnings ?? [] };
}

export function normalizedDefinition(definition: RulesetDefinition): string {
	return JSON.stringify(sortObject(definition));
}

function sortObject(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(sortObject);
	if (value && typeof value === 'object') {
		return Object.fromEntries(
			Object.entries(value as Record<string, unknown>)
				.filter(([, item]) => item !== undefined)
				.sort(([left], [right]) => left.localeCompare(right))
				.map(([key, item]) => [key, sortObject(item)])
		);
	}
	return value;
}

export function recoveryKey(rulesetId: string) {
	return `social-games-hoster:ruleset-working-copy:${rulesetId}`;
}

export function serializeRecovery(record: Omit<RecoveryRecord, 'version' | 'timestamp'>): string {
	return JSON.stringify({ ...record, version: 1, timestamp: new Date().toISOString() });
}

export function parseRecovery(value: string | null): RecoveryRecord | null {
	if (!value) return null;
	try {
		const parsed = JSON.parse(value) as RecoveryRecord;
		if (
			parsed.version !== 1 ||
			!parsed.definition ||
			!isEditorSection(parsed.section) ||
			typeof parsed.timestamp !== 'string' ||
			!parsed.selectedItems
		)
			return null;
		return parsed;
	} catch {
		return null;
	}
}

export function isEditorSection(value: string): value is EditorSection {
	return [
		'metadata',
		'teams',
		'roles',
		'composition',
		'phases',
		'knowledge',
		'chat',
		'achievements',
		'audio'
	].includes(value);
}

export function sectionForPath(path: string): EditorSection {
	if (path.startsWith('metadata') || path === 'schemaVersion') return 'metadata';
	if (path.startsWith('teams') || path.startsWith('categories')) return 'teams';
	if (path.startsWith('roles') || path.startsWith('abilities')) return 'roles';
	if (path.startsWith('composition')) return 'composition';
	if (path.startsWith('phases')) return 'phases';
	if (path.startsWith('knowledgeRules')) return 'knowledge';
	if (path.startsWith('chat')) return 'chat';
	if (path.startsWith('achievements')) return 'achievements';
	return 'audio';
}

export function sectionStates(
	definition: RulesetDefinition,
	report: ValidationReport
): Record<EditorSection, SectionState> {
	const hasErrors = (section: EditorSection) =>
		report.errors.some((issue) => sectionForPath(issue.path) === section);
	const started: Record<EditorSection, boolean> = {
		metadata: Boolean(definition.metadata.name.trim()),
		teams: definition.teams.length + definition.categories.length > 0,
		roles: definition.roles.length + definition.abilities.length > 0,
		composition: definition.compositionBands.length + definition.compositionModifiers.length > 0,
		phases: definition.phases.length > 0,
		knowledge: definition.knowledgeRules.length > 0,
		chat:
			definition.chat.channels.length > 0 ||
			Object.keys(definition.chat.phaseOverrides).length > 0 ||
			Boolean(
				definition.chat.defaultPolicy.general ||
				definition.chat.defaultPolicy.playerDm ||
				Object.keys(definition.chat.defaultPolicy.teams).length
			),
		achievements: definition.achievements.length > 0,
		audio: definition.audioCues.length > 0 || Boolean(definition.metadata.coverAssetKey)
	};
	const complete: Record<EditorSection, boolean> = {
		metadata:
			started.metadata &&
			definition.metadata.minPlayers >= 1 &&
			definition.metadata.maxPlayers <= 30 &&
			definition.metadata.minPlayers <= definition.metadata.maxPlayers,
		teams:
			definition.teams.length > 0 && definition.teams.every((item) => Boolean(item.name.trim())),
		roles:
			definition.roles.length > 0 &&
			definition.roles.every((item) => Boolean(item.name.trim() && item.teamId)),
		composition: definition.compositionBands.length > 0,
		phases: started.phases,
		knowledge: started.knowledge,
		chat: started.chat,
		achievements: started.achievements,
		audio: started.audio
	};
	return Object.fromEntries(
		(Object.keys(started) as EditorSection[]).map((section) => [
			section,
			hasErrors(section)
				? 'Needs attention'
				: complete[section]
					? 'Complete'
					: started[section]
						? 'In progress'
						: 'Not started'
		])
	) as Record<EditorSection, SectionState>;
}

export function nextRequiredSection(
	definition: RulesetDefinition,
	report: ValidationReport
): EditorSection {
	const states = sectionStates(definition, report);
	return requiredSections.find((section) => states[section] !== 'Complete') ?? 'metadata';
}

export function itemNameForIssue(
	definition: RulesetDefinition,
	issue: ValidationIssue
): string | undefined {
	const match =
		/^(teams|categories|abilities|roles|phases|compositionBands|compositionModifiers|knowledgeRules|achievements|audioCues|chat\.channels)\[(\d+)\]/.exec(
			issue.path
		);
	if (!match) return undefined;
	const index = Number(match[2]);
	if (match[1] === 'chat.channels') return definition.chat.channels[index]?.name;
	if (match[1] === 'knowledgeRules') return `Knowledge rule ${index + 1}`;
	if (match[1] === 'compositionBands') {
		const band = definition.compositionBands[index];
		return band ? `${band.minPlayers}–${band.maxPlayers} players` : undefined;
	}
	if (match[1] === 'compositionModifiers') return `Conditional change ${index + 1}`;
	const collection = definition[match[1] as keyof RulesetDefinition] as Array<{ name?: string }>;
	return collection?.[index]?.name;
}

export function humanIssueLocation(
	definition: RulesetDefinition,
	issue: ValidationIssue,
	labels: Record<EditorSection, string>
): string {
	const section = sectionForPath(issue.path);
	const item = itemNameForIssue(definition, issue);
	return item ? `${labels[section]} → ${item}` : labels[section];
}

export function issueControlName(path: string): string | undefined {
	if (path === 'metadata.name') return 'name';
	if (path.startsWith('metadata'))
		return path.includes('maxPlayers') ? 'maximum-players' : 'minimum-players';

	const slot = /^compositionBands\[(\d+)\]\.slots\[(\d+)\](?:\.(.+))?$/.exec(path);
	if (slot) {
		const [, bandIndex, slotIndex, field = 'selector'] = slot;
		if (field === 'count') return `slot-count-${bandIndex}-${slotIndex}`;
		if (field === 'label') return `slot-label-${bandIndex}-${slotIndex}`;
		const selectorField = /^selector(?:\.(roleIds|teamIds|categoryIds|tags))?$/.exec(field);
		if (selectorField) {
			const suffix = {
				roleIds: 'roles',
				teamIds: 'teams',
				categoryIds: 'categories',
				tags: 'tags'
			}[selectorField[1] ?? ''];
			return suffix
				? `slot-selector-${bandIndex}-${slotIndex}-${suffix}`
				: `slot-selector-${bandIndex}-${slotIndex}`;
		}
	}

	const knowledge =
		/^knowledgeRules\[(\d+)\]\.(viewer|target)(?:\.(roleIds|teamIds|categoryIds|tags))?$/.exec(
			path
		);
	if (knowledge) {
		const suffix = {
			roleIds: 'roles',
			teamIds: 'teams',
			categoryIds: 'categories',
			tags: 'tags'
		}[knowledge[3] ?? ''];
		return suffix
			? `knowledge-${knowledge[2]}-${knowledge[1]}-${suffix}`
			: `knowledge-${knowledge[2]}-${knowledge[1]}`;
	}

	const adjustment = /^compositionModifiers\[(\d+)\]\.slotAdjustments\[(\d+)\]/.exec(path);
	if (adjustment) return `modifier-slot-${adjustment[1]}-${adjustment[2]}`;

	const audience = /^chat\.channels\[(\d+)\]\.(readers|senders)(?:\.(roleIds|teamIds))?/.exec(path);
	if (audience) {
		const kind = audience[3] === 'teamIds' ? 'teams' : 'roles';
		return `channel-${audience[2] === 'readers' ? 'reader' : 'sender'}-${kind}-${audience[1]}`;
	}

	const match =
		/^(teams|categories|abilities|roles|phases|compositionBands|compositionModifiers|knowledgeRules|achievements|audioCues|chat\.channels)\[(\d+)\](?:\.([A-Za-z]+))?/.exec(
			path
		);
	if (!match) return undefined;
	const prefixes: Record<string, string> = {
		teams: 'team',
		categories: 'category',
		abilities: 'ability',
		roles: 'role',
		phases: 'phase',
		compositionBands: 'band',
		compositionModifiers: 'modifier',
		knowledgeRules: 'knowledge',
		achievements: 'achievement',
		audioCues: 'cue',
		'chat.channels': 'channel'
	};
	const fields: Record<string, string> = {
		teamId: 'team',
		categoryIds: 'categories',
		abilityIds: 'abilities',
		maxCopies: 'max-copies',
		activationPhaseIds: 'phases',
		suggestedDurationSeconds: 'seconds',
		audioCueId: 'audio',
		imageAssetKey: 'image',
		assetKey: 'audio-file',
		defaultAudience: 'audience',
		messageRestriction: 'message-restriction',
		senderDisplay: 'sender-display',
		whenRolePresent: 'role',
		requiresRoleIds: 'required-roles',
		excludesRoleIds: 'excluded-roles',
		name: 'name',
		points: 'points'
	};
	const defaults: Record<string, string> = {
		compositionBands: 'min',
		compositionModifiers: 'role',
		knowledgeRules: 'reveal'
	};
	return `${prefixes[match[1]]}-${fields[match[3] ?? ''] ?? defaults[match[1]] ?? 'name'}-${match[2]}`;
}

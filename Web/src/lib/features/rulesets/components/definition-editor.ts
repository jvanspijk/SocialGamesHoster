import type { RulesetDefinition, RulesetSelector } from '$lib/api/types';
import type { EditorSection } from '../editor-state';

export type DefinitionEditorSection =
	'teams' | 'roles' | 'phases' | 'composition' | 'knowledge' | 'chat' | 'achievements' | 'audio';

export type AssetOption = { assetKey: string; kind: 'image' | 'audio' };

export function nextID(prefix: string, used: string[]) {
	let candidate: string;
	do {
		candidate = `${prefix}_${crypto.randomUUID()}`;
	} while (used.includes(candidate));
	return candidate;
}

export function removeAt<T>(items: T[], index: number) {
	items.splice(index, 1);
}

export function moveByID<T extends { id: string }>(items: T[], id: string, direction: -1 | 1) {
	const index = items.findIndex((item) => item.id === id);
	const destination = index + direction;
	if (index < 0 || destination < 0 || destination >= items.length) return;
	[items[index], items[destination]] = [items[destination], items[index]];
}

export function duplicateByID<T extends { id: string; name?: string }>(
	items: T[],
	id: string,
	prefix: string
): T | undefined {
	const index = items.findIndex((item) => item.id === id);
	if (index < 0) return undefined;
	const duplicate = JSON.parse(JSON.stringify(items[index])) as T;
	duplicate.id = nextID(
		prefix,
		items.map((item) => item.id)
	);
	if (duplicate.name !== undefined) duplicate.name = `${duplicate.name || 'Unnamed'} copy`;
	items.splice(index + 1, 0, duplicate);
	return duplicate;
}

export function blankSelector(): RulesetSelector {
	return { roleIds: [], teamIds: [], categoryIds: [], tags: [] };
}

export type Usage = { label: string; section: EditorSection; itemId?: string };

export function incomingReferences(
	definition: RulesetDefinition,
	kind: 'team' | 'category' | 'ability' | 'role' | 'phase' | 'slot' | 'audioCue',
	id: string
): Usage[] {
	const usages: Usage[] = [];
	const addSelectors = (
		section: EditorSection,
		label: string,
		itemId: string | undefined,
		selectors: RulesetSelector[]
	) => {
		if (
			selectors.some((selector) =>
				kind === 'team'
					? selector.teamIds.includes(id)
					: kind === 'category'
						? selector.categoryIds.includes(id)
						: kind === 'role'
							? selector.roleIds.includes(id)
							: false
			)
		)
			usages.push({ section, label, itemId });
	};
	if (kind === 'team') {
		definition.roles
			.filter((role) => role.teamId === id)
			.forEach((role) =>
				usages.push({ section: 'roles', itemId: role.id, label: `Role · ${role.name}` })
			);
		if (definition.chat.defaultPolicy.teams[id])
			usages.push({ section: 'chat', label: 'Chat · Team room' });
		if (Object.values(definition.chat.phaseOverrides).some((override) => override.teams?.[id]))
			usages.push({ section: 'chat', label: 'Chat · Phase changes' });
		definition.chat.channels
			.filter((channel) => [...channel.readerTeamIds, ...channel.senderTeamIds].includes(id))
			.forEach((channel) =>
				usages.push({
					section: 'chat',
					itemId: channel.id,
					label: `Chat channel · ${channel.name}`
				})
			);
	}
	if (kind === 'category')
		definition.roles
			.filter((role) => role.categoryIds.includes(id))
			.forEach((role) =>
				usages.push({ section: 'roles', itemId: role.id, label: `Role · ${role.name}` })
			);
	if (kind === 'ability')
		definition.roles
			.filter((role) => role.abilityIds.includes(id))
			.forEach((role) =>
				usages.push({ section: 'roles', itemId: role.id, label: `Role · ${role.name}` })
			);
	if (kind === 'role') {
		definition.compositionModifiers
			.filter(
				(modifier) =>
					modifier.whenRolePresent === id ||
					modifier.requiresRoleIds.includes(id) ||
					modifier.excludesRoleIds.includes(id)
			)
			.forEach((modifier, index) =>
				usages.push({
					section: 'composition',
					itemId: modifier.id,
					label: `Player setup · Conditional change ${index + 1}`
				})
			);
		definition.chat.channels
			.filter((channel) => [...channel.readerRoleIds, ...channel.senderRoleIds].includes(id))
			.forEach((channel) =>
				usages.push({
					section: 'chat',
					itemId: channel.id,
					label: `Chat channel · ${channel.name}`
				})
			);
	}
	if (kind === 'phase') {
		definition.abilities
			.filter((ability) => ability.activationPhaseIds?.includes(id))
			.forEach((ability) =>
				usages.push({ section: 'roles', itemId: ability.id, label: `Ability · ${ability.name}` })
			);
		if (definition.chat.phaseOverrides[id])
			usages.push({ section: 'chat', label: 'Chat · Phase changes' });
		definition.chat.channels
			.filter((channel) => channel.phaseOverrides[id])
			.forEach((channel) =>
				usages.push({
					section: 'chat',
					itemId: channel.id,
					label: `Chat channel · ${channel.name}`
				})
			);
	}
	if (kind === 'slot')
		definition.compositionModifiers
			.filter((modifier) => modifier.slotAdjustments.some((adjustment) => adjustment.slotId === id))
			.forEach((modifier, index) =>
				usages.push({
					section: 'composition',
					itemId: modifier.id,
					label: `Conditional change ${index + 1}`
				})
			);
	if (kind === 'audioCue')
		definition.phases
			.filter((phase) => phase.audioCueId === id)
			.forEach((phase) =>
				usages.push({ section: 'phases', itemId: phase.id, label: `Phase · ${phase.name}` })
			);
	for (const [index, rule] of definition.knowledgeRules.entries())
		addSelectors('knowledge', `Information rule ${index + 1}`, undefined, [
			rule.viewer,
			rule.target
		]);
	for (const band of definition.compositionBands)
		for (const slot of band.slots)
			addSelectors('composition', `Player setup · ${slot.label}`, band.id, [slot.selector]);
	return usages;
}

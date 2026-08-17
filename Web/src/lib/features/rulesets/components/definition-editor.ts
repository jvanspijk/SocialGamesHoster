import type { RulesetSelector } from '$lib/api/types';

export type DefinitionEditorSection =
	'teams' | 'roles' | 'phases' | 'composition' | 'knowledge' | 'chat' | 'achievements' | 'audio';

export type AssetOption = { assetKey: string; kind: 'image' | 'audio' };

export function nextID(prefix: string, used: string[]) {
	let candidate = '';
	do {
		candidate = `${prefix}_${crypto.randomUUID()}`;
	} while (used.includes(candidate));
	return candidate;
}

export function removeAt<T>(items: T[], index: number) {
	items.splice(index, 1);
}

export function blankSelector(): RulesetSelector {
	return { roleIds: [], teamIds: [], categoryIds: [], tags: [] };
}

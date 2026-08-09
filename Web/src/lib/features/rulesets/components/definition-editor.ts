import type { RulesetSelector } from '$lib/api/types';

export type DefinitionEditorSection =
	| 'teams'
	| 'roles'
	| 'phases'
	| 'composition'
	| 'knowledge'
	| 'chat'
	| 'achievements'
	| 'audio';

export type AssetOption = { assetKey: string; kind: 'image' | 'audio' };

export function nextID(prefix: string, used: string[]) {
	let number = used.length + 1;
	let candidate = `${prefix}_${number}`;
	while (used.includes(candidate)) {
		number += 1;
		candidate = `${prefix}_${number}`;
	}
	return candidate;
}

export function removeAt<T>(items: T[], index: number) {
	items.splice(index, 1);
}

export function blankSelector(): RulesetSelector {
	return { roleIds: [], teamIds: [], categoryIds: [], tags: [] };
}

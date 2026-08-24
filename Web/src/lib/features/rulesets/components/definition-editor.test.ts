import { describe, expect, it } from 'vitest';
import type { RulesetDefinition } from '$lib/api/types';
import { assetUsages } from './definition-editor';

const definition: RulesetDefinition = {
	schemaVersion: 1,
	metadata: {
		name: 'Media test',
		description: '',
		minPlayers: 3,
		maxPlayers: 8,
		coverAssetKey: 'portrait'
	},
	teams: [{ id: 'team', name: 'Villagers', description: '', imageAssetKey: 'portrait' }],
	categories: [],
	abilities: [],
	roles: [],
	phases: [
		{ id: 'night', name: 'Night', description: '', order: 1, startsRound: true, audioCueId: 'bell' }
	],
	knowledgeRules: [],
	compositionBands: [],
	compositionModifiers: [],
	chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
	achievements: [],
	audioCues: [{ id: 'bell', name: 'Bell', assetKey: 'bell-file', defaultAudience: 'all' }],
	assetAccessibility: {}
};

describe('assetUsages', () => {
	it('classifies direct image usages in author-facing language', () => {
		expect(assetUsages(definition, 'portrait')).toEqual([
			{ label: 'Ruleset cover', section: 'metadata' },
			{ label: 'Team · Villagers', section: 'teams', itemId: 'team' }
		]);
	});

	it('links audio through its cue to the consuming phase', () => {
		expect(assetUsages(definition, 'bell-file')).toEqual([
			{ label: 'Audio cue · Bell → Phase · Night', section: 'phases', itemId: 'night' }
		]);
	});
});

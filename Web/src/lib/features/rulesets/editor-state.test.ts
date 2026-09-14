import { describe, expect, it, vi } from 'vitest';
import type { RulesetDefinition } from '$lib/api/types';
import {
	copyDefinition,
	humanIssueLocation,
	itemTargetForIssue,
	issueControlName,
	nextRequiredSection,
	normalizedDefinition,
	parseRecovery,
	sectionForPath,
	sectionStates,
	serializeRecovery
} from './editor-state';

function definition(): RulesetDefinition {
	return {
		schemaVersion: 1,
		metadata: { name: 'Night game', description: '', minPlayers: 4, maxPlayers: 8 },
		teams: [{ id: 'team_a', name: 'Village', description: '' }],
		categories: [],
		abilities: [],
		roles: [
			{
				id: 'role_a',
				name: 'Villager',
				description: '',
				teamId: 'team_a',
				categoryIds: [],
				tags: [],
				abilityIds: [],
				winCondition: '',
				maxCopies: 8
			}
		],
		phases: [],
		knowledgeRules: [],
		chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
		achievements: [],
		audioCues: [],
		assetAccessibility: {}
	};
}

describe('ruleset editor state', () => {
	it('normalizes object keys without hiding meaningful collection order', () => {
		const left = definition();
		const right = JSON.parse(JSON.stringify(left)) as RulesetDefinition;
		right.metadata = { maxPlayers: 8, minPlayers: 4, description: '', name: 'Night game' };
		expect(normalizedDefinition(left)).toBe(normalizedDefinition(right));
		right.teams.push({ id: 'team_b', name: 'Wolves', description: '' });
		expect(normalizedDefinition(left)).not.toBe(normalizedDefinition(right));
	});

	it('loads omitted selector lists without mutating saved data or creating edits', () => {
		const saved = definition();
		saved.knowledgeRules = [
			{
				viewer: {
					roleIds: ['role_a']
				} as RulesetDefinition['knowledgeRules'][number]['viewer'],
				target: {} as RulesetDefinition['knowledgeRules'][number]['target'],
				reveal: []
			}
		];
		const original = JSON.stringify(saved);
		const working = copyDefinition(saved);
		expect(working.knowledgeRules[0].viewer).toEqual({
			roleIds: ['role_a'],
			teamIds: [],
			categoryIds: [],
			tags: []
		});
		expect(JSON.stringify(saved)).toBe(original);
		expect(normalizedDefinition(working)).toBe(normalizedDefinition(saved));
		working.knowledgeRules[0].viewer.roleIds = [];
		expect(normalizedDefinition(working)).not.toBe(normalizedDefinition(saved));
	});

	it('does not restore empty media defaults as edits, but detects removing an image', () => {
		const saved = definition();
		const restored = copyDefinition(saved);
		restored.metadata.coverAssetKey = '';
		restored.teams[0].imageAssetKey = '';
		restored.roles[0].imageAssetKey = '';
		expect(normalizedDefinition(restored)).toBe(normalizedDefinition(saved));
		saved.metadata.coverAssetKey = 'cover';
		expect(normalizedDefinition(restored)).not.toBe(normalizedDefinition(saved));
	});

	it('round-trips recovery and rejects malformed records', () => {
		vi.useFakeTimers();
		vi.setSystemTime(new Date('2026-08-19T12:00:00Z'));
		const encoded = serializeRecovery({
			definition: definition(),
			section: 'roles',
			selectedItems: { roles: 'role_a' }
		});
		expect(parseRecovery(encoded)).toMatchObject({
			version: 2,
			section: 'roles',
			selectedItems: { roles: 'role_a' },
			timestamp: '2026-08-19T12:00:00.000Z'
		});
		expect(parseRecovery('{broken')).toBeNull();
		vi.useRealTimers();
	});

	it('restores recovery records from the former media route into assets', () => {
		const recovered = {
			version: 2,
			definition: definition(),
			section: 'media',
			selectedItems: {},
			timestamp: '2026-08-19T12:00:00.000Z'
		};

		expect(parseRecovery(JSON.stringify(recovered))?.section).toBe('assets');
	});

	it('selects the next incomplete required section and derives attention states', () => {
		const value = definition();
		value.roles = [];
		const report = { errors: [{ path: 'roles', message: 'Add at least one role.' }], warnings: [] };
		expect(nextRequiredSection(value, report)).toBe('roles');
		expect(sectionStates(value, report)).toMatchObject({
			metadata: 'Complete',
			teams: 'Complete',
			roles: 'Needs attention',
			phases: 'Not started'
		});
	});

	it('maps validation paths to author-facing destinations', () => {
		const issue = {
			path: 'knowledgeRules[0].viewer',
			message: 'Choose at least one matching role.'
		};
		expect(sectionForPath(issue.path)).toBe('knowledge');
		expect(
			humanIssueLocation(definition(), issue, {
				metadata: 'Basics',
				teams: 'Teams',
				roles: 'Roles and abilities',
				phases: 'Game flow',
				knowledge: 'Information rules',
				chat: 'Chat',
				achievements: 'Rewards',
				assets: 'Assets'
			})
		).toBe('Information rules → Knowledge rule 1');
	});

	it.each([
		['knowledgeRules[0].reveal', 'knowledge-reveal-0'],
		['knowledgeRules[1].viewer.teamIds', 'knowledge-viewer-1-teams'],
		['knowledgeRules[1].target.roleIds', 'knowledge-target-1-roles'],
		['chat.channels[2].readers.teamIds', 'channel-reader-teams-2'],
		['chat.channels[2].senders', 'channel-sender-roles-2']
	])('maps nested validation path %s to control %s', (path, control) => {
		expect(issueControlName(path)).toBe(control);
	});

	it.each([
		['categories[0].name', 'categories', 'category_a'],
		['abilities[0].name', 'abilities', 'ability_a']
	])('finds the correct collection item for %s', (path, key, id) => {
		const value = definition();
		value.categories = [{ id: 'category_a', name: 'Support', description: '' }];
		value.abilities = [{ id: 'ability_a', name: 'Inspect', description: '' }];
		expect(itemTargetForIssue(value, { path, message: 'Fix this.' })).toEqual({ key, id });
	});
});

import { describe, expect, it, vi } from 'vitest';
import type { RulesetDefinition } from '$lib/api/types';
import {
	humanIssueLocation,
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
		compositionBands: [{ id: 'band_a', minPlayers: 4, maxPlayers: 8, slots: [] }],
		compositionModifiers: [],
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

	it('selects the next incomplete required section and derives attention states', () => {
		const value = definition();
		value.roles = [];
		const report = { errors: [{ path: 'roles', message: 'Add at least one role.' }], warnings: [] };
		expect(nextRequiredSection(value, report)).toBe('roles');
		expect(sectionStates(value, report)).toMatchObject({
			metadata: 'Complete',
			teams: 'Complete',
			roles: 'Needs attention',
			composition: 'Complete',
			phases: 'Not started'
		});
	});

	it('maps validation paths to author-facing destinations', () => {
		const issue = {
			path: 'compositionBands[0].slots[0].selector',
			message: 'Choose at least one matching role.'
		};
		expect(sectionForPath(issue.path)).toBe('composition');
		expect(
			humanIssueLocation(definition(), issue, {
				metadata: 'Basics',
				teams: 'Teams',
				roles: 'Roles and abilities',
				composition: 'Player setup',
				phases: 'Game flow',
				knowledge: 'Information rules',
				chat: 'Chat',
				achievements: 'Rewards',
				audio: 'Media'
			})
		).toBe('Player setup → 4–8 players');
	});

	it.each([
		['knowledgeRules[0].reveal', 'knowledge-reveal-0'],
		['knowledgeRules[1].viewer.teamIds', 'knowledge-viewer-1-teams'],
		['knowledgeRules[1].target.roleIds', 'knowledge-target-1-roles'],
		['compositionBands[0].slots[2].count', 'slot-count-0-2'],
		['compositionBands[0].slots[2].selector', 'slot-selector-0-2'],
		['compositionBands[0].slots[2].selector.categoryIds', 'slot-selector-0-2-categories'],
		['compositionModifiers[1].slotAdjustments[3]', 'modifier-slot-1-3'],
		['chat.channels[2].readers.teamIds', 'channel-reader-teams-2'],
		['chat.channels[2].senders', 'channel-sender-roles-2']
	])('maps nested validation path %s to control %s', (path, control) => {
		expect(issueControlName(path)).toBe(control);
	});
});

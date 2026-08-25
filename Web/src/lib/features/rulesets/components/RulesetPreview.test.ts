import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { RulesetDefinition, RulesetPreviewRequest } from '$lib/api/types';
import RulesetPreview from './RulesetPreview.svelte';

afterEach(cleanup);

const definition: RulesetDefinition = {
	schemaVersion: 1,
	metadata: { name: 'Night game', description: '', minPlayers: 3, maxPlayers: 8 },
	teams: [{ id: 'team_private', name: 'Village', description: '' }],
	categories: [],
	abilities: [],
	roles: [
		{
			id: 'role_private',
			name: 'Villager',
			description: 'Find the threat.',
			teamId: 'team_private',
			categoryIds: [],
			tags: [],
			abilityIds: [],
			winCondition: 'Keep the village safe.',
			maxCopies: 8
		}
	],
	phases: [{ id: 'phase_private', name: 'Night', description: '', order: 1, startsRound: true }],
	knowledgeRules: [],
	compositionBands: [],
	compositionModifiers: [],
	chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
	achievements: [],
	audioCues: [],
	assetAccessibility: {}
};

describe('RulesetPreview', () => {
	it('labels unsaved previews and lets authors switch focused preview modes by display name', async () => {
		const onresult = vi.fn();
		const loadPreview = vi.fn(async (request: RulesetPreviewRequest) => {
			if (request.mode === 'role') {
				return {
					mode: 'role' as const,
					role: {
						name: 'Villager',
						description: 'Find the threat.',
						teamName: 'Village',
						winCondition: 'Keep the village safe.',
						abilities: []
					}
				};
			}
			return { mode: request.mode, empty: true, message: `${request.mode} preview` };
		});
		render(RulesetPreview, {
			props: {
				open: true,
				close: vi.fn(),
				definition,
				assets: [],
				dirty: true,
				loadPreview,
				onresult
			}
		});

		expect(screen.getByText('Previewing unsaved working changes')).toBeVisible();
		await waitFor(() => expect(screen.getByRole('heading', { name: 'Villager' })).toBeVisible());
		expect(onresult).toHaveBeenCalledWith(expect.objectContaining({ mode: 'role' }));
		expect(screen.getByRole('option', { name: 'Villager' })).toBeVisible();
		expect(screen.queryByText('role_private')).not.toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Player setup' }));
		await waitFor(() =>
			expect(loadPreview).toHaveBeenLastCalledWith(expect.objectContaining({ mode: 'composition' }))
		);
		expect(screen.getByRole('spinbutton', { name: 'Player count' })).toBeVisible();
	});

	it('renders selected media in each consuming game context', async () => {
		const loadPreview = vi.fn(async (request: RulesetPreviewRequest) => {
			if (request.mode !== 'media')
				return { mode: request.mode, empty: true, message: `${request.mode} preview` };
			return {
				mode: 'media' as const,
				media: {
					displayName: 'Village art',
					accessibilityText: 'A village at night',
					kind: 'image' as const,
					preview: '/api/app/v1/rulesets/preview/media'
				},
				contexts: [
					{
						kind: 'cover' as const,
						label: 'Ruleset cover',
						title: 'Night game',
						description: 'Find the threat.',
						detail: '3–8 players'
					},
					{
						kind: 'team' as const,
						label: 'Team',
						title: 'Village',
						description: 'The village side.'
					}
				]
			};
		});
		render(RulesetPreview, {
			props: {
				open: true,
				close: vi.fn(),
				definition,
				assets: [
					{
						assetKey: 'village-art',
						displayName: 'Village art',
						accessibilityText: 'A village at night',
						kind: 'image',
						mimeType: 'image/png',
						checksum: 'checksum',
						metadata: {},
						preview: '/api/app/v1/rulesets/preview/media',
						staged: false,
						usages: []
					}
				],
				dirty: false,
				loadPreview
			}
		});

		await fireEvent.click(screen.getByRole('button', { name: 'Media' }));
		await waitFor(() => expect(screen.getByRole('heading', { name: 'Night game' })).toBeVisible());
		expect(screen.getByRole('heading', { name: 'Village' })).toBeVisible();
		expect(screen.getByText('3–8 players')).toBeVisible();
		expect(screen.getByText('The village side.')).toBeVisible();
	});
});

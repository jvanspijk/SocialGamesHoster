import { cleanup, render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { RulesetDefinition } from '$lib/api/types';
import MediaLibrary from './MediaLibrary.svelte';

afterEach(cleanup);

const definition: RulesetDefinition = {
	schemaVersion: 1,
	metadata: { name: 'Test ruleset', description: '', minPlayers: 3, maxPlayers: 12 },
	teams: [],
	categories: [],
	abilities: [],
	roles: [],
	phases: [],
	knowledgeRules: [],
	compositionBands: [],
	compositionModifiers: [],
	chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
	achievements: [],
	audioCues: [],
	assetAccessibility: {}
};

const assets = [
	{
		assetKey: 'image-internal',
		displayName: 'Village portrait',
		accessibilityText: 'A village at dusk',
		kind: 'image' as const,
		mimeType: 'image/png',
		checksum: 'image-checksum',
		metadata: { width: 10, height: 10 },
		preview: '/image-preview',
		staged: false,
		usages: []
	},
	{
		assetKey: 'audio-internal',
		displayName: 'Night bell',
		accessibilityText: 'A bell rings',
		kind: 'audio' as const,
		mimeType: 'audio/ogg',
		checksum: 'audio-checksum',
		metadata: { durationSeconds: 2 },
		preview: '/audio-preview',
		staged: false,
		usages: []
	}
];

describe('MediaLibrary', () => {
	it('uses native keyboard-operable buttons for media selection', async () => {
		const user = userEvent.setup();
		render(MediaLibrary, {
			props: {
				definition,
				assets,
				media: { upload: vi.fn(), update: vi.fn(), remove: vi.fn() },
				onnavigate: vi.fn()
			}
		});

		expect(screen.queryByRole('listbox')).not.toBeInTheDocument();
		const village = screen.getByRole('button', { name: /Village portrait/ });
		village.focus();
		await user.keyboard('{Enter}');
		expect(screen.getByRole('heading', { name: 'Village portrait' })).toBeVisible();
		expect(village).toHaveAttribute('aria-current', 'true');

		await user.tab();
		const bell = screen.getByRole('button', { name: /Night bell/ });
		expect(bell).toHaveFocus();
		await user.keyboard('{Enter}');
		expect(screen.getByRole('heading', { name: 'Night bell' })).toBeVisible();
	});
});

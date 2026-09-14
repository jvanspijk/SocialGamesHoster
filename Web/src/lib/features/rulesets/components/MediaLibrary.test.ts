import { cleanup, render, screen, within } from '@testing-library/svelte';
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
	it('offers matching upload flows for empty image and sound sections', async () => {
		const user = userEvent.setup();
		const upload = vi.fn().mockResolvedValueOnce(assets[0]).mockResolvedValueOnce(assets[1]);
		render(MediaLibrary, {
			props: {
				definition,
				assets: [],
				media: { upload, update: vi.fn(), remove: vi.fn() },
				onnavigate: vi.fn()
			}
		});

		const images = screen.getByRole('region', { name: 'Images' });
		const sounds = screen.getByRole('region', { name: 'Sounds' });
		expect(within(images).getByRole('button', { name: 'Upload image' })).toBeVisible();
		expect(within(sounds).getByRole('button', { name: 'Upload sound' })).toBeVisible();
		expect(screen.queryByLabelText(/audience/i)).not.toBeInTheDocument();

		const image = new File(['image'], 'village.png', { type: 'image/png' });
		await user.upload(screen.getByLabelText('Choose image to upload'), image);
		const sound = new File(['sound'], 'bell.ogg', { type: 'audio/ogg' });
		await user.upload(screen.getByLabelText('Choose sound to upload'), sound);

		expect(upload).toHaveBeenNthCalledWith(1, image, 'image', 'village.png', '');
		expect(upload).toHaveBeenNthCalledWith(2, sound, 'audio', 'bell.ogg', '');
	});

	it('uses one search field to filter images and sounds together', async () => {
		const user = userEvent.setup();
		render(MediaLibrary, {
			props: {
				definition,
				assets,
				media: { upload: vi.fn(), update: vi.fn(), remove: vi.fn() },
				onnavigate: vi.fn()
			}
		});

		await user.type(screen.getByRole('textbox', { name: 'Search assets' }), 'night');

		expect(screen.queryByRole('button', { name: /Village portrait/ })).not.toBeInTheDocument();
		expect(screen.getByRole('button', { name: /Night bell/ })).toBeVisible();
		expect(screen.getByText('No matching images.')).toBeVisible();
	});

	it('uses native keyboard-operable buttons for asset selection', async () => {
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
		expect(screen.getAllByRole('heading', { name: 'Item details' })).toHaveLength(1);
		expect(screen.getByRole('heading', { name: 'Village portrait' })).toBeVisible();
		expect(village).toHaveAttribute('aria-current', 'true');
		await user.click(village);
		expect(screen.queryByRole('heading', { name: 'Village portrait' })).not.toBeInTheDocument();
		expect(village).not.toHaveAttribute('aria-current');

		const bell = screen.getByRole('button', { name: /Night bell/ });
		bell.focus();
		await user.keyboard('{Enter}');
		expect(screen.getByRole('heading', { name: 'Night bell' })).toBeVisible();
	});
});

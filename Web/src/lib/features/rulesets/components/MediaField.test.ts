import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import MediaField from './MediaField.svelte';

afterEach(cleanup);

const asset = {
	assetKey: 'asset_internal_123',
	displayName: 'Village portrait',
	accessibilityText: 'A village at dusk',
	kind: 'image' as const,
	mimeType: 'image/png',
	checksum: 'checksum',
	metadata: { width: 10, height: 10 },
	preview: '/preview',
	staged: false,
	usages: []
};

describe('MediaField', () => {
	it('selects No image when the usage has no asset', () => {
		render(MediaField, {
			props: {
				label: 'Ruleset cover',
				name: 'ruleset-cover',
				kind: 'image',
				assets: [],
				media: { upload: vi.fn(), update: vi.fn(), remove: vi.fn() }
			}
		});

		expect(screen.getByRole('combobox', { name: /Ruleset cover/ })).toHaveValue('');
		expect((screen.getByRole('option', { name: 'No image' }) as HTMLOptionElement).selected).toBe(
			true
		);
	});

	it('falls back to No image when a saved cover asset no longer exists', async () => {
		render(MediaField, {
			props: {
				label: 'Ruleset cover',
				name: 'ruleset-cover',
				kind: 'image',
				value: 'missing-cover',
				assets: [asset],
				media: { upload: vi.fn(), update: vi.fn(), remove: vi.fn() }
			}
		});

		await waitFor(() =>
			expect(screen.getByRole('combobox', { name: /Ruleset cover/ })).toHaveValue('')
		);
		expect((screen.getByRole('option', { name: 'No image' }) as HTMLOptionElement).selected).toBe(
			true
		);
	});

	it('uses display names for media choices without exposing internal references', () => {
		render(MediaField, {
			props: {
				label: 'Team image',
				name: 'team-image',
				kind: 'image',
				assets: [asset],
				media: { upload: vi.fn(), update: vi.fn(), remove: vi.fn() }
			}
		});

		expect(screen.getByRole('option', { name: 'Village portrait' })).toBeVisible();
		expect(screen.queryByText('asset_internal_123')).not.toBeInTheDocument();
	});

	it('collects reusable accessibility metadata when uploading at a usage', async () => {
		const upload = vi.fn().mockResolvedValue(asset);
		const { container } = render(MediaField, {
			props: {
				label: 'Team image',
				name: 'team-image',
				kind: 'image',
				assets: [],
				media: { upload, update: vi.fn(), remove: vi.fn() }
			}
		});
		await fireEvent.input(screen.getByRole('textbox', { name: /Asset name/ }), {
			target: { value: 'Town square' }
		});
		await fireEvent.input(screen.getByRole('textbox', { name: /Image description/ }), {
			target: { value: 'Players gathered in a square' }
		});
		const file = new File(['png'], 'square.png', { type: 'image/png' });
		await fireEvent.change(container.querySelector('input[type="file"]')!, {
			target: { files: [file] }
		});
		await waitFor(() =>
			expect(upload).toHaveBeenCalledWith(
				file,
				'image',
				'Town square',
				'Players gathered in a square'
			)
		);
	});

	it('redirects upload creation when the usage provides a media-page action', async () => {
		const onuploadnew = vi.fn();
		render(MediaField, {
			props: {
				label: 'Ruleset cover',
				name: 'ruleset-cover',
				kind: 'image',
				compact: true,
				assets: [],
				media: { upload: vi.fn(), update: vi.fn(), remove: vi.fn() },
				onuploadnew
			}
		});

		await fireEvent.click(screen.getByRole('button', { name: 'Upload new' }));

		expect(onuploadnew).toHaveBeenCalledOnce();
	});

	it('replaces only this usage with a newly uploaded Asset', async () => {
		const replacement = { ...asset, assetKey: 'replacement' };
		const upload = vi.fn().mockResolvedValue(replacement);
		const { container } = render(MediaField, {
			props: {
				label: 'Ruleset cover',
				name: 'ruleset-cover',
				kind: 'image',
				compact: true,
				value: asset.assetKey,
				assets: [asset],
				media: { upload, update: vi.fn(), remove: vi.fn() }
			}
		});

		expect(screen.getByRole('button', { name: 'Replace' })).toBeVisible();
		expect(screen.queryByRole('button', { name: 'Upload new' })).not.toBeInTheDocument();
		expect(
			screen.queryByRole('button', { name: 'Remove from this usage' })
		).not.toBeInTheDocument();

		const file = new File(['png'], 'new-cover.png', { type: 'image/png' });
		await fireEvent.change(container.querySelector('input[type="file"]')!, {
			target: { files: [file] }
		});

		await waitFor(() => expect(upload).toHaveBeenCalledWith(file, 'image', 'new-cover.png', ''));
	});
});

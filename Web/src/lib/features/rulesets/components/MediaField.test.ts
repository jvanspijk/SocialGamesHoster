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
		await fireEvent.input(screen.getByRole('textbox', { name: /Media name/ }), {
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
});

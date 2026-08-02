import { cleanup, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Media from './Media.svelte';

afterEach(() => {
	cleanup();
	vi.restoreAllMocks();
});

describe('Media', () => {
	it('renders an accessible image from a prepared URL', () => {
		render(Media, { props: { source: 'blob:portrait', kind: 'image', alt: 'Player portrait' } });

		expect(screen.getByRole('img', { name: 'Player portrait' })).toHaveAttribute(
			'src',
			'blob:portrait'
		);
	});

	it('loads protected-ready media through an explicit loader and reports failures', async () => {
		let resolve: (source: Blob) => void;
		const loader = vi.fn(
			() =>
				new Promise<Blob>((complete) => {
					resolve = complete;
				})
		);
		vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:loaded-audio');
		render(Media, { props: { loader, kind: 'audio', autoplay: true } });

		expect(screen.getByText('Loading media…')).toBeVisible();
		resolve!(new Blob(['sound']));
		await waitFor(() => expect(document.querySelector('audio')).toBeInTheDocument());
		expect(loader).toHaveBeenCalledOnce();
		expect(document.querySelector('audio')).toHaveAttribute('autoplay');
		expect(document.querySelector('audio')).toHaveAttribute('preload', 'auto');

		render(Media, {
			props: { loader: () => Promise.reject(new Error('unavailable')), kind: 'image' }
		});
		await waitFor(() => expect(screen.getByText('Media unavailable')).toBeVisible());
	});

	it('releases object URLs when prepared blob sources change or unmount', async () => {
		const revoke = vi.spyOn(URL, 'revokeObjectURL');
		vi.spyOn(URL, 'createObjectURL')
			.mockReturnValueOnce('blob:first')
			.mockReturnValueOnce('blob:second');
		const view = render(Media, { props: { source: new Blob(['image']), kind: 'image', alt: '' } });

		await waitFor(() =>
			expect(view.container.querySelector('img')).toHaveAttribute('src', 'blob:first')
		);
		await view.rerender({ source: new Blob(['replacement']), kind: 'image', alt: '' });
		await waitFor(() =>
			expect(view.container.querySelector('img')).toHaveAttribute('src', 'blob:second')
		);
		expect(revoke).toHaveBeenCalledWith('blob:first');
		view.unmount();
		expect(revoke).toHaveBeenCalledWith('blob:second');
	});
});

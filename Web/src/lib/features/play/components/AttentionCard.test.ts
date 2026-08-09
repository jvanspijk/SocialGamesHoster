import { cleanup, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { AnnouncementAttentionItem, FutureEventAttentionItem } from '$lib/api/types';
import AttentionCard from './AttentionCard.svelte';

vi.mock('$lib/api/client', () => ({
	fetchBlob: vi.fn().mockResolvedValue(new Blob(['audio']))
}));

const item = {
	id: 'announcement-1',
	kind: 'announcement',
	senderLabel: 'Host',
	content: 'Listen carefully.',
	createdAt: '2026-07-28T12:00:00Z',
	audio: { url: '/api/app/v1/assets/audio', alternative: 'A short rising tone.' }
} satisfies AnnouncementAttentionItem;

afterEach(() => {
	cleanup();
	vi.restoreAllMocks();
});

describe('AttentionCard', () => {
	it('autoplays announcement audio without exposing media controls', async () => {
		vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:announcement-audio');

		render(AttentionCard, {
			props: { item, position: 1, total: 1, acknowledge: vi.fn() }
		});

		await waitFor(() => expect(document.querySelector('audio')).toBeInTheDocument());
		const audio = document.querySelector('audio');
		expect(audio).toHaveAttribute('autoplay');
		expect(audio).not.toHaveAttribute('controls');
		expect(audio).toHaveAttribute('preload', 'auto');
		expect(screen.getByText('A short rising tone.')).toBeInTheDocument();
	});

	it('presents the queue, announcement text, image alternative, and acknowledgement action', async () => {
		vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:announcement-image');
		const acknowledge = vi.fn();
		const announcement = {
			...item,
			image: { url: '/api/app/v1/assets/image', description: 'A red wax seal.' }
		};

		render(AttentionCard, {
			props: { item: announcement, position: 2, total: 3, acknowledge }
		});

		expect(screen.getByLabelText('Announcement 2 of 3')).toBeInTheDocument();
		expect(screen.getByText('2 of 3')).toBeInTheDocument();
		expect(screen.getByText('Announcement from Host')).toBeInTheDocument();
		expect(screen.getByText('Listen carefully.')).toBeInTheDocument();
		expect(screen.getByText('A red wax seal.')).toBeInTheDocument();
		await screen.findByAltText('A red wax seal.');

		screen.getByRole('button', { name: 'Acknowledge' }).click();
		expect(acknowledge).toHaveBeenCalledOnce();
	});

	it('disables a busy acknowledgement action', () => {
		render(AttentionCard, {
			props: { item, position: 1, total: 1, acknowledge: vi.fn(), busy: true }
		});

		expect(screen.getByRole('button', { name: /Acknowledging/ })).toBeDisabled();
	});

	it('shows an unavailable message for unsupported attention items', () => {
		const unsupported = {
			id: 'event-1',
			kind: 'event',
			createdAt: '2026-07-28T12:00:00Z'
		} satisfies FutureEventAttentionItem;

		render(AttentionCard, {
			props: { item: unsupported, position: 1, total: 1, acknowledge: vi.fn() }
		});

		expect(
			screen.getByText('This event type is not available in this version.')
		).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Acknowledge' })).not.toBeInTheDocument();
	});
});

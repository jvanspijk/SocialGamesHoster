import { render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { AnnouncementAttentionItem } from '$lib/api/types';
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
});

import { cleanup, render, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import TimerDisplay from './TimerDisplay.svelte';

const mocks = vi.hoisted(() => ({ api: vi.fn() }));

vi.mock('$lib/api/client', () => ({ api: mocks.api }));

afterEach(() => {
	cleanup();
	mocks.api.mockReset();
});

describe('TimerDisplay', () => {
	it('refreshes when the player view receives a newer game revision', async () => {
		mocks.api.mockResolvedValue({
			status: 'running',
			totalMs: 60_000,
			remainingMs: 60_000,
			revision: 1,
			serverTime: '2026-09-16T12:00:00Z'
		});

		const rendered = render(TimerDisplay, { props: { gameId: 'game-1', revision: 1 } });
		await waitFor(() => expect(mocks.api).toHaveBeenCalledTimes(1));

		rendered.rerender({ gameId: 'game-1', revision: 2 });
		await waitFor(() => expect(mocks.api).toHaveBeenCalledTimes(2));
		expect(mocks.api).toHaveBeenLastCalledWith('/games/game-1/timer');
	});
});

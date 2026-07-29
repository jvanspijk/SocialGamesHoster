import { cleanup, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import ChatApp from './ChatApp.svelte';
import type { Room } from '$lib/api/types';

const api = vi.hoisted(() => vi.fn());

vi.mock('$lib/api/client', () => ({
	api,
	jsonBody: (value: unknown) => ({ body: value }),
	pb: { realtime: { subscribe: vi.fn().mockResolvedValue(() => undefined) } }
}));

vi.mock('$lib/state/auth.svelte', () => ({ auth: { actor: { id: 'player-1' } } }));
vi.mock('$lib/state/toasts.svelte', () => ({ toasts: { error: vi.fn(), success: vi.fn() } }));

afterEach(cleanup);

function deferred<T>() {
	let resolve: (value: T) => void;
	const promise = new Promise<T>((settle) => {
		resolve = settle;
	});
	return { promise, resolve: resolve! };
}

const lookoutReport: Room = {
	id: 'lookout-report',
	key: 'custom:lookout_report',
	kind: 'custom',
	label: 'Lookout Report',
	playersCanPost: true,
	readable: true,
	sendable: true,
	messageRestriction: 'emoji_only',
	latestMessage: null
};

describe('ChatApp', () => {
	it('keeps the newest room permissions when a phase changes during loading', async () => {
		const observationRooms = deferred<Room[]>();
		api
			.mockImplementationOnce(() => observationRooms.promise)
			.mockResolvedValueOnce([lookoutReport]);

		const { rerender } = render(ChatApp, {
			props: {
				gameId: 'game-1',
				policyRevision: 'running:observation',
				selectRoom: vi.fn()
			}
		});

		await waitFor(() => expect(api).toHaveBeenCalledTimes(1));
		await rerender({ policyRevision: 'running:report_window' });
		await waitFor(() => expect(api).toHaveBeenCalledTimes(2));
		await waitFor(() => expect(screen.getByText('Lookout Report')).toBeInTheDocument());

		observationRooms.resolve([]);
		await Promise.resolve();

		expect(screen.getByText('Lookout Report')).toBeInTheDocument();
	});
});

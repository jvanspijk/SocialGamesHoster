import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import PlayerChatPage from './PlayerChatPage.svelte';

const mocks = vi.hoisted(() => ({
	api: vi.fn(),
	goto: vi.fn(),
	player: null as unknown,
	toastError: vi.fn()
}));

vi.mock('$app/navigation', () => ({ goto: mocks.goto }));
vi.mock('$app/paths', () => ({ resolve: (path: string) => path }));
vi.mock('$lib/api/client', () => ({
	api: mocks.api,
	jsonBody: (value: unknown) => ({ body: value }),
	pb: { realtime: { subscribe: vi.fn().mockResolvedValue(() => undefined) } }
}));
vi.mock('$lib/state/auth.svelte', () => ({ auth: { actor: { id: 'profile-1' } } }));
vi.mock('$lib/state/toasts.svelte', () => ({
	toasts: { error: mocks.toastError, success: vi.fn() }
}));
vi.mock('$lib/state/game.svelte', () => ({
	gameState: {
		get player() {
			return mocks.player;
		}
	}
}));

beforeEach(() => {
	mocks.api.mockReset();
	mocks.goto.mockReset();
	mocks.toastError.mockReset();
	mocks.api.mockResolvedValue([]);
	mocks.player = {
		game: { id: 'game-1', status: 'running', phaseKey: 'day' },
		party: [
			{
				id: 'participant-1',
				profileId: 'profile-1',
				displayName: 'You',
				gameAlias: '',
				seatNumber: 1,
				status: 'active'
			},
			{
				id: 'participant-2',
				profileId: 'profile-2',
				displayName: 'Rowan',
				gameAlias: '',
				seatNumber: 2,
				status: 'active'
			}
		]
	};
});

afterEach(cleanup);

describe('PlayerChatPage', () => {
	it('creates or opens a player room before navigating to it', async () => {
		mocks.api.mockResolvedValueOnce([]).mockResolvedValueOnce({ id: 'room-2' });
		render(PlayerChatPage);
		await fireEvent.click(screen.getAllByRole('button', { name: 'New message' })[0]);
		await fireEvent.click(screen.getByRole('button', { name: /Rowan.*Seat 2/ }));

		await waitFor(() =>
			expect(mocks.api).toHaveBeenCalledWith('/games/game-1/rooms/player-dm', {
				method: 'POST',
				body: { participantId: 'participant-2' }
			})
		);
		expect(mocks.goto).toHaveBeenCalledWith('/play/chat/room-2');
		expect(screen.queryByRole('button', { name: /You.*Seat 1/ })).not.toBeInTheDocument();
	});
});

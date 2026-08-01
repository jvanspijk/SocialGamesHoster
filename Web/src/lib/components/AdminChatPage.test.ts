import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import AdminChatPage from './AdminChatPage.svelte';

const mocks = vi.hoisted(() => ({
	api: vi.fn(),
	goto: vi.fn(),
	admin: null as unknown
}));

vi.mock('$app/navigation', () => ({ goto: mocks.goto }));
vi.mock('$app/paths', () => ({ resolve: (path: string) => path }));
vi.mock('$lib/api/client', () => ({
	api: mocks.api,
	jsonBody: (value: unknown) => ({ body: value }),
	pb: { realtime: { subscribe: vi.fn().mockResolvedValue(() => undefined) } }
}));
vi.mock('$lib/state/auth.svelte', () => ({ auth: { actor: { id: 'profile-1' } } }));
vi.mock('$lib/state/toasts.svelte', () => ({ toasts: { error: vi.fn(), success: vi.fn() } }));
vi.mock('$lib/state/game.svelte', () => ({
	gameState: {
		get admin() {
			return mocks.admin;
		}
	}
}));

beforeEach(() => {
	mocks.api.mockReset();
	mocks.goto.mockReset();
	mocks.api.mockResolvedValue([]);
	mocks.admin = {
		game: { id: 'game-1', status: 'running', phaseKey: 'day' },
		participants: [
			{
				id: 'participant-2',
				displayNameSnapshot: 'Rowan',
				gameAlias: '',
				seatNumber: 2,
				status: 'active'
			},
			{
				id: 'participant-left',
				displayNameSnapshot: 'Left player',
				gameAlias: '',
				seatNumber: 3,
				status: 'left'
			}
		],
		rooms: [{ id: 'room-2', key: 'gm:participant-2' }]
	};
});

afterEach(cleanup);

describe('AdminChatPage', () => {
	it('opens the existing game-master room without creating a player room', async () => {
		render(AdminChatPage);
		await fireEvent.click(screen.getAllByRole('button', { name: 'New message' })[0]);
		await fireEvent.click(screen.getByRole('button', { name: /Rowan.*Seat 2/ }));

		await waitFor(() => expect(mocks.goto).toHaveBeenCalledWith('/admin/games/game-1/chat/room-2'));
		expect(mocks.api).toHaveBeenCalledWith('/games/game-1/rooms');
		expect(mocks.api).not.toHaveBeenCalledWith('/games/game-1/rooms/player-dm', expect.anything());
		expect(screen.queryByRole('button', { name: /Left player/ })).not.toBeInTheDocument();
	});
});

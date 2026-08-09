import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/svelte';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import AdminChatPage from './AdminChatPage.svelte';
import PlayerChatPage from './PlayerChatPage.svelte';

const mocks = vi.hoisted(() => ({
	api: vi.fn(),
	goto: vi.fn(),
	admin: null as unknown,
	player: null as unknown
}));
vi.mock('$app/navigation', () => ({ goto: mocks.goto }));
vi.mock('$app/paths', () => ({ resolve: (path: string) => path }));
vi.mock('$lib/api/client', () => ({
	api: mocks.api,
	jsonBody: (value: unknown) => ({ body: value }),
	pb: { realtime: { subscribe: vi.fn().mockResolvedValue(() => undefined) } }
}));
vi.mock('$lib/state/auth.svelte', () => ({ auth: { actor: { id: 'profile-1' } } }));
vi.mock('$lib/state/toasts.svelte', () => ({ toasts: { error: vi.fn() } }));
vi.mock('$lib/state/game.svelte', () => ({
	gameState: {
		get admin() {
			return mocks.admin;
		},
		get player() {
			return mocks.player;
		}
	}
}));
beforeEach(() => {
	mocks.api.mockReset();
	mocks.api.mockResolvedValue([]);
	mocks.goto.mockReset();
});
afterEach(cleanup);

function newMessageButton() {
	return within(screen.getByRole('heading', { name: 'Conversations' }).closest('aside')!).getByRole(
		'button',
		{ name: 'New message' }
	);
}

it('offers admin active players and opens their existing game-master conversation', async () => {
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
	render(AdminChatPage);
	await fireEvent.click(newMessageButton());
	await fireEvent.click(screen.getByRole('button', { name: /Rowan.*Seat 2/ }));
	await waitFor(() => expect(mocks.goto).toHaveBeenCalledWith('/admin/games/game-1/chat/room-2'));
	expect(screen.queryByRole('button', { name: /Left player/ })).not.toBeInTheDocument();
});

it('excludes the player and creates their direct conversation before navigating', async () => {
	mocks.api.mockResolvedValueOnce([]).mockResolvedValueOnce({ id: 'room-2' });
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
			},
			{
				id: 'participant-eliminated',
				profileId: 'profile-3',
				displayName: 'Eliminated player',
				gameAlias: '',
				seatNumber: 3,
				status: 'eliminated'
			}
		]
	};
	render(PlayerChatPage);
	await fireEvent.click(newMessageButton());
	await fireEvent.click(screen.getByRole('button', { name: /Rowan.*Seat 2/ }));
	await waitFor(() =>
		expect(mocks.api).toHaveBeenCalledWith('/games/game-1/rooms/player-dm', {
			method: 'POST',
			body: { participantId: 'participant-2' }
		})
	);
	expect(mocks.goto).toHaveBeenCalledWith('/play/chat/room-2');
	expect(screen.queryByRole('button', { name: /You.*Seat 1/ })).not.toBeInTheDocument();
	expect(
		screen.queryByRole('button', { name: /Eliminated player.*Seat 3/ })
	).not.toBeInTheDocument();
});

import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/svelte';
import { afterEach, beforeEach, expect, it, vi } from 'vitest';
import DirectMessageChatPage from './DirectMessageChatPage.svelte';

const mocks = vi.hoisted(() => ({ api: vi.fn(), goto: vi.fn() }));
vi.mock('$app/navigation', () => ({ goto: mocks.goto }));
vi.mock('$lib/api/client', () => ({
	api: mocks.api,
	jsonBody: (value: unknown) => ({ body: value }),
	pb: { realtime: { subscribe: vi.fn().mockResolvedValue(() => undefined) } }
}));
vi.mock('$lib/state/auth.svelte', () => ({ auth: { actor: { id: 'profile-1' } } }));
vi.mock('$lib/state/toasts.svelte', () => ({ toasts: { error: vi.fn(), success: vi.fn() } }));
beforeEach(() => {
	mocks.api.mockReset();
	mocks.api.mockResolvedValue([]);
});
afterEach(cleanup);

function newMessageButton() {
	return within(screen.getByRole('heading', { name: 'Conversations' }).closest('aside')!).getByRole(
		'button',
		{ name: 'New message' }
	);
}

it('opens a recipient, closes the dialog, and selects the returned conversation', async () => {
	const openConversation = vi.fn().mockResolvedValue('room-2');
	render(DirectMessageChatPage, {
		gameId: 'game-1',
		recipientEntries: [{ id: 'participant-2', label: 'Rowan', accessibleLabel: 'Rowan, Seat 2' }],
		roomPath: (roomId) => `/play/chat/${roomId}`,
		openConversation
	});
	await fireEvent.click(newMessageButton());
	await fireEvent.click(screen.getByRole('button', { name: /Rowan.*Seat 2/ }));
	await waitFor(() => expect(openConversation).toHaveBeenCalledWith('participant-2'));
	expect(mocks.goto).toHaveBeenCalledWith('/play/chat/room-2');
	await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument());
});

it('keeps the recipient dialog open when opening fails', async () => {
	render(DirectMessageChatPage, {
		gameId: 'game-1',
		recipientEntries: [{ id: 'participant-2', label: 'Rowan', accessibleLabel: 'Rowan, Seat 2' }],
		roomPath: (roomId) => `/play/chat/${roomId}`,
		openConversation: () => undefined
	});
	await fireEvent.click(newMessageButton());
	await fireEvent.click(screen.getByRole('button', { name: /Rowan.*Seat 2/ }));
	expect(screen.getByRole('dialog')).toBeInTheDocument();
});

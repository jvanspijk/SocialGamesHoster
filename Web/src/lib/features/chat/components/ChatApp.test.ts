import { cleanup, fireEvent, render, screen, waitFor, within } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ChatApp from './ChatApp.svelte';
import type { Room } from '$lib/api/types';
import { readMarkerStorageKey } from '$lib/state/chatReadMarkers';

const api = vi.hoisted(() => vi.fn());
const subscribe = vi.hoisted(() => vi.fn());

vi.mock('$lib/api/client', () => ({
	api,
	jsonBody: (value: unknown) => ({ body: value }),
	pb: { realtime: { subscribe } }
}));

vi.mock('$lib/state/auth.svelte', () => ({ auth: { actor: { id: 'player-1' } } }));
vi.mock('$lib/state/toasts.svelte', () => ({ toasts: { error: vi.fn(), success: vi.fn() } }));

beforeEach(() => {
	api.mockReset();
	subscribe.mockReset().mockResolvedValue(() => undefined);
	localStorage.clear();
	Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
		configurable: true,
		value: vi.fn()
	});
});

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

const generalRoom: Room = {
	id: 'general',
	key: 'general',
	kind: 'general',
	label: 'General',
	playersCanPost: true,
	readable: true,
	sendable: true,
	latestMessage: {
		id: 'message-1',
		createdAt: '2026-07-27T12:00:01Z',
		senderLabel: 'Alex',
		preview: 'Earlier'
	}
};

const firstMessage = {
	id: 'message-1',
	roomId: 'general',
	kind: 'chat',
	senderType: 'player',
	senderLabel: 'Alex',
	content: 'Earlier',
	deleted: false,
	createdAt: '2026-07-27T12:00:01Z'
};

const secondMessage = {
	id: 'message-2',
	roomId: 'general',
	kind: 'chat',
	senderType: 'player',
	senderLabel: 'Blake',
	content: 'Latest',
	deleted: false,
	createdAt: '2026-07-27T12:00:02Z'
};

const teamRoom: Room = {
	...generalRoom,
	id: 'team-one',
	key: 'team:one',
	kind: 'team',
	label: 'Team One',
	latestMessage: {
		id: 'team-message-1',
		createdAt: '2026-07-27T12:01:00Z',
		senderLabel: 'Casey',
		preview: 'Team update'
	}
};

const teamMessage = {
	...firstMessage,
	id: 'team-message-1',
	roomId: 'team-one',
	senderLabel: 'Casey',
	content: 'Team update',
	createdAt: '2026-07-27T12:01:00Z'
};

describe('ChatApp', () => {
	it('loads, selects, paginates, sends, removes, and marks messages through the controller', async () => {
		const olderMessage = {
			...firstMessage,
			id: 'message-0',
			content: 'Oldest',
			createdAt: '2026-07-27T11:00:00Z'
		};
		const deletedMessage = { ...secondMessage, content: '', deleted: true };
		api.mockImplementation((path: string, options?: { method?: string }) => {
			if (path === '/games/game-1/rooms') return Promise.resolve([generalRoom]);
			if (path === '/rooms/general/messages' && options?.method === 'POST') {
				return Promise.resolve({ ...secondMessage });
			}
			if (path === '/rooms/general/messages?cursor=older-cursor') {
				return Promise.resolve({ items: [olderMessage], nextCursor: '' });
			}
			if (path === '/rooms/general/messages') {
				return Promise.resolve({ items: [firstMessage], nextCursor: 'older-cursor' });
			}
			if (path === '/rooms/general/messages/message-2' && options?.method === 'DELETE') {
				return Promise.resolve(deletedMessage);
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		const selectRoom = vi.fn((roomId: string) => {
			void rendered.rerender({ selectedRoomId: roomId });
		});
		const rendered = render(ChatApp, {
			props: {
				gameId: 'game-1',
				policyRevision: 'running:day',
				selectRoom
			}
		});

		await waitFor(() =>
			expect(screen.getByRole('button', { name: 'General, New messages' })).toBeInTheDocument()
		);
		await fireEvent.click(screen.getByRole('button', { name: 'General, New messages' }));
		await waitFor(() => expect(screen.getByText('Earlier')).toBeInTheDocument());
		expect(selectRoom).toHaveBeenCalledWith('general');
		expect(screen.queryByRole('button', { name: 'General, New messages' })).not.toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Load earlier messages' }));
		await waitFor(() => expect(screen.getByText('Oldest')).toBeInTheDocument());

		const composer = screen.getByRole('textbox', { name: 'Message' });
		await fireEvent.input(composer, { target: { value: 'Latest' } });
		expect(composer).toHaveValue('Latest');
		await fireEvent.click(screen.getByRole('button', { name: 'Send' }));
		await waitFor(() => expect(screen.getByText('Latest')).toBeInTheDocument());
		expect(api).toHaveBeenCalledWith('/rooms/general/messages', {
			method: 'POST',
			body: { content: 'Latest' }
		});

		await fireEvent.click(screen.getByRole('button', { name: 'Remove message from Blake' }));
		await waitFor(() => expect(screen.getByText('Message removed')).toBeInTheDocument());
		expect(api).toHaveBeenCalledWith('/rooms/general/messages/message-2', { method: 'DELETE' });
	});

	it('keeps archived conversations read-only and hides the composer', async () => {
		const archivedRoom = {
			...generalRoom,
			id: 'archived-room',
			label: 'Archived',
			playersCanPost: false,
			sendable: false,
			latestMessage: null
		};
		api.mockImplementation((path: string) => {
			if (path === '/games/game-1/rooms') return Promise.resolve([archivedRoom]);
			if (path === '/rooms/archived-room/messages') {
				return Promise.resolve({ items: [], nextCursor: '' });
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		render(ChatApp, {
			props: {
				gameId: 'game-1',
				selectedRoomId: 'archived-room',
				policyRevision: 'archived:finished',
				archived: true,
				selectRoom: vi.fn()
			}
		});

		await waitFor(() => expect(screen.getByText('Archived chat is read-only')).toBeInTheDocument());
		expect(screen.queryByRole('textbox', { name: 'Message' })).not.toBeInTheDocument();
		await waitFor(() =>
			expect(screen.getByRole('heading', { name: 'No messages yet' })).toBeInTheDocument()
		);
	});

	it('refreshes the selected conversation when posting policy changes', async () => {
		let roomLoads = 0;
		let messageLoads = 0;
		api.mockImplementation((path: string) => {
			if (path === '/games/game-1/rooms') {
				roomLoads += 1;
				return Promise.resolve([
					{
						...generalRoom,
						playersCanPost: roomLoads === 1,
						latestMessage:
							roomLoads === 1
								? generalRoom.latestMessage
								: {
										id: secondMessage.id,
										createdAt: secondMessage.createdAt,
										senderLabel: secondMessage.senderLabel,
										preview: secondMessage.content
									}
					}
				]);
			}
			if (path === '/rooms/general/messages') {
				messageLoads += 1;
				return Promise.resolve({
					items: messageLoads === 1 ? [firstMessage] : [secondMessage, firstMessage],
					nextCursor: ''
				});
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		const { rerender } = render(ChatApp, {
			props: {
				gameId: 'game-1',
				selectedRoomId: 'general',
				policyRevision: 'running:discussion',
				selectRoom: vi.fn()
			}
		});

		const messages = await screen.findByRole('log', { name: 'Messages' });
		await waitFor(() => expect(within(messages).getByText('Earlier')).toBeVisible());
		messages.scrollTop = 77;
		await rerender({ policyRevision: 'running:night' });

		await waitFor(() => expect(messageLoads).toBe(2));
		await waitFor(() => expect(within(messages).getByText('Latest')).toBeVisible());
		expect(within(messages).getByRole('separator', { name: 'New messages' })).toBeInTheDocument();
		expect(messages.scrollTop).toBe(77);
		expect(
			JSON.parse(localStorage.getItem(readMarkerStorageKey('player-1', 'game-1')) ?? '{}')
		).toMatchObject({ general: { id: 'message-2', createdAt: secondMessage.createdAt } });
		expect(screen.getByText(/Players read-only/)).toBeInTheDocument();
	});

	it('deduplicates realtime messages and persists the newest read marker', async () => {
		let realtimeHandler: ((event: unknown) => void) | undefined;
		subscribe.mockImplementation((_topic: string, handler: (event: unknown) => void) => {
			realtimeHandler = handler;
			return Promise.resolve(() => undefined);
		});
		api.mockImplementation((path: string) => {
			if (path === '/games/game-1/rooms') return Promise.resolve([generalRoom]);
			if (path === '/rooms/general/messages') {
				return Promise.resolve({ items: [firstMessage], nextCursor: '' });
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		render(ChatApp, {
			props: {
				gameId: 'game-1',
				selectedRoomId: 'general',
				policyRevision: 'running:day',
				selectRoom: vi.fn()
			}
		});

		const messages = await screen.findByRole('log', { name: 'Messages' });
		await waitFor(() => expect(within(messages).getByText('Earlier')).toBeVisible());
		expect(realtimeHandler).toBeTypeOf('function');
		realtimeHandler?.({ kind: 'chat.message_created', payload: firstMessage });
		realtimeHandler?.({ kind: 'chat.message_created', payload: secondMessage });

		await waitFor(() => expect(within(messages).getByText('Latest')).toBeVisible());
		expect(within(messages).getAllByText('Earlier')).toHaveLength(1);
		expect(within(messages).getAllByText('Latest')).toHaveLength(1);
		expect(
			JSON.parse(localStorage.getItem(readMarkerStorageKey('player-1', 'game-1')) ?? '{}')
		).toMatchObject({ general: { id: 'message-2', createdAt: secondMessage.createdAt } });
	});

	it('places the first-unread divider from the persisted marker and advances it after reading', async () => {
		localStorage.setItem(
			readMarkerStorageKey('player-1', 'game-1'),
			JSON.stringify({ general: { id: firstMessage.id, createdAt: firstMessage.createdAt } })
		);
		api.mockImplementation((path: string) => {
			if (path === '/games/game-1/rooms') {
				return Promise.resolve([{ ...generalRoom, latestMessage: secondMessage }]);
			}
			if (path === '/rooms/general/messages') {
				return Promise.resolve({ items: [secondMessage, firstMessage], nextCursor: '' });
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		render(ChatApp, {
			props: {
				gameId: 'game-1',
				selectedRoomId: 'general',
				policyRevision: 'running:day',
				selectRoom: vi.fn()
			}
		});

		const messages = await screen.findByRole('log', { name: 'Messages' });
		const divider = await within(messages).findByRole('separator', { name: 'New messages' });
		const messageArticles = within(messages).getAllByRole('article');
		expect(divider).toHaveAttribute('id', 'unread-message-2');
		expect(messageArticles[0]).toHaveTextContent('Earlier');
		expect(divider.nextElementSibling).toBe(messageArticles[1]);
		expect(api.mock.calls.filter(([path]) => path === '/rooms/general/messages')).toHaveLength(1);
		expect(
			JSON.parse(localStorage.getItem(readMarkerStorageKey('player-1', 'game-1')) ?? '{}')
		).toMatchObject({ general: { id: 'message-2', createdAt: secondMessage.createdAt } });
	});

	it('restores each conversation scroll position after switching rooms', async () => {
		api.mockImplementation((path: string) => {
			if (path === '/games/game-1/rooms') return Promise.resolve([generalRoom, teamRoom]);
			if (path === '/rooms/general/messages') {
				return Promise.resolve({ items: [firstMessage], nextCursor: '' });
			}
			if (path === '/rooms/team-one/messages') {
				return Promise.resolve({ items: [teamMessage], nextCursor: '' });
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		const { rerender } = render(ChatApp, {
			props: {
				gameId: 'game-1',
				selectedRoomId: 'general',
				policyRevision: 'running:day',
				selectRoom: vi.fn()
			}
		});

		const messages = await screen.findByRole('log', { name: 'Messages' });
		await waitFor(() => expect(within(messages).getByText('Earlier')).toBeVisible());
		Object.defineProperty(messages, 'scrollHeight', { configurable: true, value: 500 });
		messages.scrollTop = 137;

		await rerender({ selectedRoomId: 'team-one' });
		await waitFor(() => expect(within(messages).getByText('Team update')).toBeVisible());
		messages.scrollTop = 42;

		await rerender({ selectedRoomId: 'general' });
		await waitFor(() => expect(within(messages).getByText('Earlier')).toBeVisible());
		expect(messages.scrollTop).toBe(137);
	});

	it('updates posting access through the named moderation control', async () => {
		api.mockImplementation((path: string, options?: { method?: string }) => {
			if (path === '/games/game-1/rooms') return Promise.resolve([generalRoom]);
			if (path === '/rooms/general/messages') {
				return Promise.resolve({ items: [firstMessage], nextCursor: '' });
			}
			if (path === '/rooms/general' && options?.method === 'PATCH') {
				return Promise.resolve({ ...generalRoom, playersCanPost: false });
			}
			throw new Error(`Unexpected request: ${path}`);
		});

		render(ChatApp, {
			props: {
				gameId: 'game-1',
				selectedRoomId: 'general',
				policyRevision: 'running:day',
				canModerate: true,
				selectRoom: vi.fn()
			}
		});

		const posting = await screen.findByRole('checkbox', { name: 'Players can post' });
		expect(posting).toBeChecked();
		await fireEvent.click(posting);

		await waitFor(() =>
			expect(api).toHaveBeenCalledWith('/rooms/general', {
				method: 'PATCH',
				body: { playersCanPost: false }
			})
		);
		expect(posting).not.toBeChecked();
		expect(screen.getByText(/Players read-only/)).toBeInTheDocument();
	});

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

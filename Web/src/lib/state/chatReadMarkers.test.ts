import { describe, expect, it, vi } from 'vitest';
import {
	countUnreadMessages,
	cursorIsAfter,
	hasUnreadMessages,
	readMarkerStorageKey
} from './chatReadMarkers';

describe('chat read markers', () => {
	it('namespaces device-local state by actor and game', () => {
		expect(readMarkerStorageKey('actor-a', 'game-a')).not.toBe(
			readMarkerStorageKey('actor-b', 'game-a')
		);
		expect(readMarkerStorageKey('actor-a', 'game-a')).not.toBe(
			readMarkerStorageKey('actor-a', 'game-b')
		);
	});

	it('compares the stable createdAt and id cursor', () => {
		const marker = { createdAt: '2026-07-27T12:00:00Z', id: 'message-a' };
		expect(cursorIsAfter({ createdAt: '2026-07-27T12:00:01Z', id: 'message-a' }, marker)).toBe(
			true
		);
		expect(cursorIsAfter({ createdAt: marker.createdAt, id: 'message-b' }, marker)).toBe(true);
		expect(cursorIsAfter({ createdAt: marker.createdAt, id: marker.id }, marker)).toBe(false);
		expect(cursorIsAfter(null, marker)).toBe(false);
	});

	it('reports unread messages across any conversation', () => {
		const rooms = [
			{
				id: 'general',
				latestMessage: {
					id: 'message-a',
					createdAt: '2026-07-27T12:00:00Z',
					senderLabel: 'Alex',
					preview: 'Hello'
				}
			},
			{
				id: 'dm',
				latestMessage: {
					id: 'message-b',
					createdAt: '2026-07-27T12:00:01Z',
					senderLabel: 'Blake',
					preview: 'Hi'
				}
			}
		];
		expect(hasUnreadMessages(rooms, { general: rooms[0].latestMessage })).toBe(true);
		expect(
			hasUnreadMessages(rooms, {
				general: rooms[0].latestMessage,
				dm: rooms[1].latestMessage
			})
		).toBe(false);
	});

	it('counts every unread message up to the display limit', async () => {
		const rooms = [
			{
				id: 'general',
				latestMessage: {
					id: 'message-c',
					createdAt: '2026-07-27T12:00:02Z',
					senderLabel: 'Casey',
					preview: 'Later'
				}
			}
		];
		const loadPage = vi.fn(async (_roomId: string, cursor: string) =>
			cursor
				? { items: [{ id: 'message-a', createdAt: '2026-07-27T12:00:00Z' }], nextCursor: '' }
				: {
						items: [
							{ id: 'message-c', createdAt: '2026-07-27T12:00:02Z' },
							{ id: 'message-b', createdAt: '2026-07-27T12:00:01Z' }
						],
						nextCursor: 'older'
					}
		);

		await expect(
			countUnreadMessages(
				rooms,
				{ general: { id: 'message-a', createdAt: '2026-07-27T12:00:00Z' } },
				loadPage
			)
		).resolves.toBe(2);
		expect(loadPage).toHaveBeenCalledTimes(2);
	});
});

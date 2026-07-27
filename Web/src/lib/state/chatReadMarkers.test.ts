import { describe, expect, it } from 'vitest';
import { cursorIsAfter, readMarkerStorageKey } from './chatReadMarkers';

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
});

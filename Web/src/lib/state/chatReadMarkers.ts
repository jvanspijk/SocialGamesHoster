import type { MessageSummary } from '$lib/api/types';

export type ReadMarkers = Record<string, Pick<MessageSummary, 'id' | 'createdAt'>>;

export function readMarkerStorageKey(actorId: string, gameId: string) {
	return `sgh.read.v1:${actorId}:${gameId}`;
}

export function cursorIsAfter(
	candidate: Pick<MessageSummary, 'id' | 'createdAt'> | null | undefined,
	marker: Pick<MessageSummary, 'id' | 'createdAt'> | undefined
) {
	if (!candidate) return false;
	if (!marker) return true;
	const timeDifference =
		new Date(candidate.createdAt).getTime() - new Date(marker.createdAt).getTime();
	return timeDifference > 0 || (timeDifference === 0 && candidate.id > marker.id);
}

export function readMarkers(actorId: string, gameId: string): ReadMarkers {
	try {
		return JSON.parse(localStorage.getItem(readMarkerStorageKey(actorId, gameId)) ?? '{}');
	} catch {
		return {};
	}
}

export function hasUnreadMessages(
	rooms: Array<{ id: string; latestMessage: MessageSummary | null }>,
	markers: ReadMarkers
) {
	return rooms.some((room) => cursorIsAfter(room.latestMessage, markers[room.id]));
}

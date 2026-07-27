import type { MessageCursor } from '$lib/api/types';

export function readMarkerStorageKey(actorId: string, gameId: string) {
	return `sgh.read.v1:${actorId}:${gameId}`;
}

export function cursorIsAfter(
	candidate: MessageCursor | null | undefined,
	marker: MessageCursor | undefined
) {
	if (!candidate) return false;
	if (!marker) return true;
	const timeDifference =
		new Date(candidate.createdAt).getTime() - new Date(marker.createdAt).getTime();
	return timeDifference > 0 || (timeDifference === 0 && candidate.id > marker.id);
}

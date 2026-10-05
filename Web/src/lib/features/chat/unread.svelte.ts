import { api, jsonBody } from '$lib/api/client';
import {
	chatReadMarkersChanged,
	readMarkers,
	readMarkerStorageKey
} from '$lib/state/chatReadMarkers';

export interface UnreadCounts {
	counts: Record<string, number>;
	total: number;
}

export function unreadChatLabel(total: number) {
	return total
		? `Chat, ${total === 99 ? '99 or more' : total} unread ${total === 1 ? 'message' : 'messages'}`
		: 'Chat';
}

// Each shell owns one controller. Coalesce triggers and allow only one request in flight.
export function createChatUnreadController() {
	let result = $state<UnreadCounts>({ counts: {}, total: 0 });
	let promise = $state.raw<Promise<UnreadCounts> | null>(null);
	let actorId = '';
	let gameId = '';
	let revision = -1;
	let generation = 0;
	let dirty = false;
	let running = false;
	let disposed = false;
	let ready = false;

	async function drain() {
		if (running || disposed || !ready) return result;
		running = true;
		try {
			while (dirty && gameId && actorId && !disposed && ready) {
				dirty = false;
				const request = generation;
				try {
					const loaded = await api<UnreadCounts>(`/games/${gameId}/unread-counts`, {
						method: 'POST',
						...jsonBody({ markers: readMarkers(actorId, gameId) })
					});
					if (!disposed && request === generation) result = loaded;
				} catch {
					// Keep the last badge value; a later event retries the background read.
				}
			}
		} finally {
			running = false;
		}
		return result;
	}

	function refresh() {
		if (disposed) return;
		generation += 1;
		dirty = true;
		queueMicrotask(() => {
			if (!running && !disposed && ready && actorId && gameId) promise = drain();
		});
	}

	function storageChanged(event: StorageEvent) {
		if (event.key === null || event.key === readMarkerStorageKey(actorId, gameId)) refresh();
	}

	return {
		get promise() {
			return promise;
		},
		get total() {
			return result.total;
		},
		get counts() {
			return result.counts;
		},
		get label() {
			return unreadChatLabel(result.total);
		},
		refresh,
		setContext(actor: string, game: string, nextRevision: number, nextReady = true) {
			if (actor === actorId && game === gameId && nextRevision === revision && nextReady === ready)
				return;
			if (actor !== actorId || game !== gameId) {
				result = { counts: {}, total: 0 };
				if (!running) promise = null;
			}
			actorId = actor;
			gameId = game;
			revision = nextRevision;
			ready = nextReady;
			refresh();
		},
		start() {
			window.addEventListener(chatReadMarkersChanged, refresh);
			window.addEventListener('storage', storageChanged);
			return () => {
				disposed = true;
				generation += 1;
				window.removeEventListener(chatReadMarkersChanged, refresh);
				window.removeEventListener('storage', storageChanged);
			};
		}
	};
}

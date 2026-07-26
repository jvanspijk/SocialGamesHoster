import { api, pb } from '$lib/api/client';
import type { AdminGameView, PlayerGameView, RealtimeEnvelope } from '$lib/api/types';
import { connection } from './connection.svelte';

let playerView = $state<PlayerGameView | null>(null);
let adminView = $state<AdminGameView | null>(null);
let lastEventId = $state('');

async function refreshPlayer() {
	playerView = await api<PlayerGameView>('/games/live/player-view');
	return playerView;
}

async function refreshAdmin(gameId: string) {
	adminView = await api<AdminGameView>(`/games/${gameId}/admin-view`);
	return adminView;
}

async function subscribe(topic: string, refresh: () => Promise<unknown>) {
	const unsubscribe = await pb.realtime.subscribe(topic, async (message) => {
		const envelope = message as unknown as RealtimeEnvelope;
		if (!envelope.eventId || envelope.eventId === lastEventId) return;
		lastEventId = envelope.eventId;
		const currentRevision = playerView?.game.revision ?? adminView?.game.revision ?? 0;
		if (envelope.revision && envelope.revision > currentRevision + 1) {
			connection.set('reconnecting');
		}
		await refresh();
		connection.set('connected');
	});
	connection.set('connected');
	return unsubscribe;
}

export const gameState = {
	get player() {
		return playerView;
	},
	get admin() {
		return adminView;
	},
	refreshPlayer,
	refreshAdmin,
	subscribe,
	clear() {
		playerView = null;
		adminView = null;
		lastEventId = '';
	}
};

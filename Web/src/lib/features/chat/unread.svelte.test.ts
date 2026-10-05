import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createChatUnreadController, unreadChatLabel, type UnreadCounts } from './unread.svelte';
import { chatReadMarkersChanged, readMarkerStorageKey } from '$lib/state/chatReadMarkers';

const api = vi.hoisted(() => vi.fn());
vi.mock('$lib/api/client', () => ({ api, jsonBody: (value: unknown) => ({ body: value }) }));

const stop: Array<() => void> = [];
beforeEach(() => {
	api.mockReset();
	localStorage.clear();
});
afterEach(() => {
	for (const dispose of stop.splice(0)) dispose();
});

function deferred() {
	let resolve!: (value: UnreadCounts) => void;
	const promise = new Promise<UnreadCounts>((settle) => {
		resolve = settle;
	});
	return { promise, resolve };
}

function controller() {
	const unread = createChatUnreadController();
	stop.push(unread.start());
	return unread;
}

describe('shared chat unread counts', () => {
	it('uses device read markers and a single count request without loading history', async () => {
		const markers = { general: { id: 'message', createdAt: '2026-10-05T12:00:00Z' } };
		localStorage.setItem(readMarkerStorageKey('actor', 'game'), JSON.stringify(markers));
		api.mockResolvedValue({ counts: { general: 3 }, total: 3 });
		const unread = controller();
		unread.setContext('actor', 'game', 1);
		await vi.waitFor(() => expect(unread.total).toBe(3));
		expect(api).toHaveBeenCalledExactlyOnceWith('/games/game/unread-counts', {
			method: 'POST',
			body: { markers }
		});
		expect(unread.label).toBe('Chat, 3 unread messages');
	});

	it('coalesces bursts and discards results overtaken by a read-marker change', async () => {
		const first = deferred();
		const second = deferred();
		api.mockReturnValueOnce(first.promise).mockReturnValueOnce(second.promise);
		const unread = controller();
		unread.setContext('actor', 'game', 1);
		unread.refresh();
		await vi.waitFor(() => expect(api).toHaveBeenCalledTimes(1));
		for (let i = 0; i < 5; i++) window.dispatchEvent(new Event(chatReadMarkersChanged));
		expect(api).toHaveBeenCalledTimes(1);
		first.resolve({ counts: { general: 5 }, total: 5 });
		await vi.waitFor(() => expect(api).toHaveBeenCalledTimes(2));
		expect(unread.total).toBe(0);
		second.resolve({ counts: { general: 0 }, total: 0 });
		await vi.waitFor(() => expect(unread.total).toBe(0));
	});

	it('clears counts on account/game changes and ignores an old in-flight result', async () => {
		const old = deferred();
		api.mockReturnValueOnce(old.promise).mockResolvedValueOnce({ counts: { room: 2 }, total: 2 });
		const unread = controller();
		unread.setContext('actor', 'game', 1);
		await vi.waitFor(() => expect(api).toHaveBeenCalledTimes(1));
		unread.setContext('other', 'other-game', 1);
		old.resolve({ counts: { private: 9 }, total: 9 });
		await vi.waitFor(() => expect(unread.total).toBe(2));
		expect(unread.counts).toEqual({ room: 2 });
		expect(api.mock.calls[1][0]).toBe('/games/other-game/unread-counts');
	});

	it('refreshes for this device namespace when another tab changes markers', async () => {
		api.mockResolvedValue({ counts: {}, total: 0 });
		const unread = controller();
		unread.setContext('actor', 'game', 1);
		await vi.waitFor(() => expect(api).toHaveBeenCalledTimes(1));
		window.dispatchEvent(
			new StorageEvent('storage', { key: readMarkerStorageKey('other', 'game') })
		);
		await Promise.resolve();
		expect(api).toHaveBeenCalledTimes(1);
		window.dispatchEvent(
			new StorageEvent('storage', { key: readMarkerStorageKey('actor', 'game') })
		);
		await vi.waitFor(() => expect(api).toHaveBeenCalledTimes(2));
	});

	it('preserves the badge on transient failures and retries on later events', async () => {
		api
			.mockResolvedValueOnce({ counts: { room: 2 }, total: 2 })
			.mockRejectedValueOnce(new Error('offline'))
			.mockResolvedValueOnce({ counts: {}, total: 0 });
		const unread = controller();
		unread.setContext('actor', 'game', 1);
		await vi.waitFor(() => expect(unread.total).toBe(2));
		unread.refresh();
		await vi.waitFor(() => expect(api).toHaveBeenCalledTimes(2));
		expect(unread.total).toBe(2);
		unread.refresh();
		await vi.waitFor(() => expect(unread.total).toBe(0));
	});

	it('gives player and admin badges the same accessible wording', () => {
		expect(unreadChatLabel(0)).toBe('Chat');
		expect(unreadChatLabel(1)).toBe('Chat, 1 unread message');
		expect(unreadChatLabel(99)).toBe('Chat, 99 or more unread messages');
	});
});

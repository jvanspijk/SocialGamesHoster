import { createServer, type ServerResponse } from 'node:http';
import type { Page, Route } from '@playwright/test';
import type { Room } from '../src/lib/api/types';
import { expect, test } from './fixtures';

function room(id = 'general', kind: Room['kind'] = 'general'): Room {
	return {
		id,
		key: id,
		kind,
		label: id === 'general' ? 'General' : 'Private conversation',
		playersCanPost: true,
		readable: true,
		sendable: true,
		latestMessage: null
	};
}

async function openChat(
	page: Page,
	rooms: Room[],
	options: {
		admin?: boolean;
		patch?: (route: Route, room: Room) => Promise<void>;
		realtimeURL?: string;
		onSubscriptions?: (topics: string[]) => void;
	} = {}
) {
	const admin = options.admin ?? true;
	await page.addInitScript((isAdmin) => {
		const token = `test.${btoa(JSON.stringify({ exp: 4_000_000_000 }))}.test`;
		localStorage.setItem(
			'pocketbase_auth',
			JSON.stringify({
				token,
				record: {
					id: 'actor',
					type: isAdmin ? 'game_masters' : 'player_profiles',
					displayName: 'Chat tester'
				}
			})
		);
	}, admin);
	await page.route('**/api/realtime', async (route) => {
		if (route.request().method() !== 'GET') {
			options.onSubscriptions?.(route.request().postDataJSON().subscriptions);
			return route.fulfill({ json: {} });
		}
		if (options.realtimeURL) return route.continue();
		return route.fulfill({
			contentType: 'text/event-stream',
			body: 'id: moderation\nevent: PB_CONNECT\ndata: {}\n\n'
		});
	});
	const commands: Array<{ id: string; playersCanPost: boolean }> = [];
	await page.route('**/api/app/**', async (route) => {
		const request = route.request();
		const path = new URL(request.url()).pathname;
		const game = {
			id: 'moderation',
			name: 'Chat test',
			status: 'running',
			phaseKey: 'day',
			revision: 1
		};
		if (path.endsWith('/admin-view') || path.endsWith('/player-view'))
			return route.fulfill({
				json: {
					game,
					rooms,
					participants: [],
					participant: { id: 'participant' },
					party: [],
					ruleset: {},
					attentionItems: [],
					assets: []
				}
			});
		if (path.endsWith('/unread-counts')) return route.fulfill({ json: { counts: {}, total: 0 } });
		if (path.endsWith('/profile-requests')) return route.fulfill({ json: [] });
		if (path.endsWith('/rooms')) return route.fulfill({ json: rooms });
		if (path.endsWith('/messages')) return route.fulfill({ json: { items: [], nextCursor: '' } });
		if (request.method() === 'PATCH') {
			const id = path.split('/').at(-1)!;
			const updated = { ...rooms.find((item) => item.id === id)!, ...request.postDataJSON() };
			commands.push({ id, playersCanPost: updated.playersCanPost });
			if (options.patch) return options.patch(route, updated);
			return route.fulfill({ json: updated });
		}
		return route.fulfill({ json: { version: 'test', needsOwner: false } });
	});
	await page.goto(
		`${options.realtimeURL ?? ''}${admin ? '/admin/games/moderation/chat' : '/play/chat'}`
	);
	await expect(page.getByRole('searchbox', { name: 'Search conversations' })).toBeVisible();
	return commands;
}

test('locks and unlocks the selected channel', async ({ page }) => {
	const commands = await openChat(page, [room()]);
	await page.getByRole('button', { name: 'General', exact: true }).click();
	await page.getByRole('button', { name: 'Lock channel', exact: true }).click();
	await expect(page.getByRole('button', { name: 'Unlock channel', exact: true })).toBeEnabled();
	await expect(page.getByText('· Players read-only', { exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Unlock channel', exact: true }).click();
	await expect(page.getByRole('button', { name: 'Lock channel', exact: true })).toBeEnabled();
	expect(commands).toEqual([
		{ id: 'general', playersCanPost: false },
		{ id: 'general', playersCanPost: true }
	]);
});

for (const kind of ['gm_dm', 'player_dm'] as const) {
	test(`locks and unlocks all channels with only a ${kind} conversation`, async ({ page }) => {
		const commands = await openChat(page, [room('direct', kind)]);
		await page.getByRole('button', { name: 'Lock all channels', exact: true }).click();
		await expect(
			page.getByRole('button', { name: 'Private conversation, Players read-only', exact: true })
		).toBeEnabled();
		await page.getByRole('button', { name: 'Unlock all channels', exact: true }).click();
		await expect(
			page.getByRole('button', { name: 'Private conversation', exact: true })
		).toBeEnabled();
		expect(commands).toEqual([
			{ id: 'direct', playersCanPost: false },
			{ id: 'direct', playersCanPost: true }
		]);
	});
}

test('keeps successful locks after a failure and retries only unlocked channels', async ({
	page
}) => {
	let failTeam = true;
	let release!: () => void;
	const pending = new Promise<void>((resolve) => {
		release = resolve;
	});
	const team = { ...room('team', 'team'), label: 'Team' };
	const commands = await openChat(page, [room(), team], {
		patch: async (route, updated) => {
			if (updated.id === 'general') await pending;
			if (updated.id === 'team' && failTeam)
				return route.fulfill({
					status: 503,
					json: { code: 'test.unavailable', message: 'Team unavailable' }
				});
			await route.fulfill({ json: updated });
		}
	});
	const lock = page.getByRole('button', { name: 'Lock all channels', exact: true });
	await lock.click();
	await expect(lock).toBeDisabled();
	release();
	await expect(page.getByText('Team unavailable', { exact: true })).toBeVisible();
	await expect(
		page.getByRole('button', { name: 'General, Players read-only', exact: true })
	).toBeEnabled();
	await expect(page.getByRole('button', { name: 'Team', exact: true })).toBeEnabled();
	await expect(
		page.getByText('Players cannot post. They can still read messages.', { exact: true })
	).toHaveCount(0);
	failTeam = false;
	await lock.click();
	await expect(
		page.getByRole('button', { name: 'Unlock all channels', exact: true })
	).toBeEnabled();
	expect(commands).toEqual([
		{ id: 'general', playersCanPost: false },
		{ id: 'team', playersCanPost: false },
		{ id: 'team', playersCanPost: false }
	]);
});

for (const admin of [true, false]) {
	test(`updates ${admin ? 'admin' : 'player'} channel locks from realtime and keeps them selectable`, async ({
		page,
		baseURL
	}) => {
		let connection: ServerResponse | undefined;
		let subscriptions: string[] = [];
		// A real SSE stream exercises the browser's native EventSource and PocketBase delivery.
		const server = createServer(async (request, response) => {
			if (request.url !== '/api/realtime') {
				try {
					const upstream = await fetch(`${baseURL}${request.url}`);
					response.writeHead(upstream.status, {
						'Content-Type': upstream.headers.get('content-type') ?? 'application/octet-stream'
					});
					response.end(Buffer.from(await upstream.arrayBuffer()));
				} catch {
					response.writeHead(502).end();
				}
				return;
			}
			response.writeHead(200, { 'Content-Type': 'text/event-stream' });
			connection = response;
			response.write('id: moderation\nevent: PB_CONNECT\ndata: {}\n\n');
		});
		await new Promise<void>((resolve) => server.listen(0, '127.0.0.1', resolve));
		const address = server.address();
		if (!address || typeof address === 'string') throw new Error('Cannot allocate SSE port');
		try {
			await openChat(page, [room()], {
				admin,
				realtimeURL: `http://127.0.0.1:${address.port}`,
				onSubscriptions: (topics) => {
					subscriptions = topics;
				}
			});
			// The shell subscribes to rooms too; wait for chat to load before sending an update.
			await expect(page.getByRole('button', { name: 'General', exact: true })).toBeEnabled();
			await expect.poll(() => subscriptions).toContain('room:general');
			connection!.write(
				`event: room:general\ndata: ${JSON.stringify({ kind: 'chat.room_updated', payload: { id: 'general', playersCanPost: false } })}\n\n`
			);
			const locked = page.getByRole('button', { name: 'General, Players read-only', exact: true });
			await expect(locked).toBeEnabled();
			await locked.click();
			await expect(
				page.getByRole('heading', { name: 'No messages yet', exact: true })
			).toBeVisible();
			if (!admin)
				await expect(
					page.getByRole('button', { name: 'Unlock all channels', exact: true })
				).toHaveCount(0);
			connection!.write(
				`event: room:general\ndata: ${JSON.stringify({ kind: 'chat.room_updated', payload: { id: 'general', playersCanPost: true } })}\n\n`
			);
			await expect(page.getByRole('button', { name: 'General', exact: true })).toBeEnabled();
		} finally {
			await page.close();
			connection?.end();
			server.closeAllConnections();
			await new Promise<void>((resolve, reject) =>
				server.close((error) => (error ? reject(error) : resolve()))
			);
		}
	});
}

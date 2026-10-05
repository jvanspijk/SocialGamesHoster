import { expect, test } from './fixtures';

for (const admin of [true, false]) {
	test(`${admin ? 'admin' : 'player'} chat counts unread messages without downloading history and preserves the rail`, async ({
		page
	}) => {
		await page.setViewportSize({ width: 390, height: 844 });
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
		const rooms = ['general', 'team'].map((id) => ({
			id,
			key: id,
			kind: 'general',
			label: id === 'general' ? 'General' : 'Team',
			playersCanPost: true,
			readable: true,
			sendable: true,
			latestMessage: {
				id: `${id}-message`,
				createdAt: '2026-10-05T12:00:00Z',
				senderLabel: 'Alex',
				preview: `${id} hello`
			}
		}));
		let roomLoads = 0;
		let messageLoads = 0;
		let countLoads = 0;
		let releaseCounts!: () => void;
		const countsReady = new Promise<void>((resolve) => {
			releaseCounts = resolve;
		});
		await page.route('**/api/realtime', (route) =>
			route.request().method() === 'GET'
				? route.fulfill({
						contentType: 'text/event-stream',
						body: 'id: chat-test\nevent: PB_CONNECT\ndata: {}\n\n'
					})
				: route.fulfill({ json: {} })
		);
		await page.route('**/api/app/**', async (route) => {
			const path = new URL(route.request().url()).pathname;
			const game = {
				id: 'chat-test',
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
			if (path.endsWith('/unread-counts')) {
				countLoads++;
				await countsReady;
				const markers = route.request().postDataJSON().markers;
				const counts = Object.fromEntries(
					rooms.map((room) => [room.id, markers[room.id] ? 0 : room.id === 'general' ? 2 : 1])
				);
				return route.fulfill({
					json: { counts, total: Object.values(counts).reduce((sum, count) => sum + count, 0) }
				});
			}
			if (path.endsWith('/rooms')) {
				roomLoads++;
				return route.fulfill({ json: rooms });
			}
			if (path.endsWith('/messages')) {
				messageLoads++;
				const id = path.split('/').at(-2);
				return route.fulfill({
					json: {
						items: [
							{
								id: `${id}-message`,
								roomId: id,
								createdAt: '2026-10-05T12:00:00Z',
								senderLabel: 'Alex',
								content: `${id} hello`,
								deleted: false
							}
						],
						nextCursor: ''
					}
				});
			}
			if (path.endsWith('/profile-requests')) return route.fulfill({ json: [] });
			return route.fulfill({ json: { version: 'test', needsOwner: false } });
		});
		const route = admin ? '/admin/games/chat-test/chat' : '/play/chat';
		await page.goto(route);
		await expect(page.getByRole('searchbox', { name: 'Search conversations' })).toBeVisible();
		await expect(page.getByRole('link', { name: 'Chat', exact: true })).toBeVisible();
		await expect(
			page.getByRole('link', { name: 'Chat, 3 unread messages', exact: true })
		).toHaveCount(0);
		await expect.poll(() => countLoads).toBe(1);
		releaseCounts();
		await expect(
			page.getByRole('link', { name: 'Chat, 3 unread messages', exact: true })
		).toBeVisible();
		await expect(
			page.getByRole('button', { name: 'General, New messages', exact: true })
		).toBeVisible();
		expect(countLoads).toBe(1);
		expect(messageLoads).toBe(0);
		expect(roomLoads).toBe(1);
		await page.getByRole('searchbox', { name: 'Search conversations' }).fill('General');
		await page.getByRole('button', { name: 'General, New messages', exact: true }).click();
		await expect(
			page.getByRole('log', { name: 'Messages' }).getByText('general hello')
		).toBeVisible();
		await expect(
			page.getByRole('link', { name: 'Chat, 1 unread message', exact: true })
		).toBeVisible();
		await page.getByRole('button', { name: 'Back to conversations' }).click();
		await expect(page.getByRole('searchbox', { name: 'Search conversations' })).toHaveValue(
			'General'
		);
		expect(messageLoads).toBe(1);
		expect(roomLoads).toBe(1);
	});
}

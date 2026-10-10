import type { Page } from '@playwright/test';
import type { AdminAttentionSummary, AdminGameView, Game } from '../src/lib/api/types';
import { expect, test } from './fixtures';

function gameView(status: Game['status'] = 'running'): AdminGameView {
	return {
		game: {
			id: 'actionstest',
			name: 'Friday night',
			status,
			rulesetVersion: 'v1',
			joiningOpen: status === 'lobby',
			rolesVisible: true,
			roleVisibilityRevision: 1,
			revision: 1,
			roundNumber: 3,
			phaseKey: 'discussion'
		},
		timer: {
			status: 'paused',
			totalMs: 300_000,
			remainingMs: 240_000,
			revision: 1,
			serverTime: new Date().toISOString()
		},
		ruleset: {
			schemaVersion: 1,
			metadata: { name: 'Test', description: '', minPlayers: 2, maxPlayers: 30 },
			phases: [
				{
					id: 'discussion',
					name: 'Discussion',
					description: 'Discuss the round.',
					order: 1,
					startsRound: false
				},
				{
					id: 'voting',
					name: 'Voting',
					description: 'Cast your vote.',
					order: 2,
					startsRound: true,
					suggestedDurationSeconds: 60
				}
			],
			roles: [
				{
					id: 'scout',
					name: 'Scout',
					teamId: 'blue',
					description: '',
					categoryIds: [],
					tags: [],
					abilityIds: [],
					winCondition: ''
				}
			],
			teams: [{ id: 'blue', name: 'Blue Team', description: '' }],
			categories: [],
			abilities: [],
			knowledgeRules: [],
			audioCues: [],
			chat: { defaultPolicy: { teams: {} }, phaseOverrides: {}, channels: [] },
			achievements: [
				{
					id: 'hero',
					name: 'Hero',
					description: 'Help another player.',
					points: 1,
					hiddenUntilGameCompleted: false
				}
			]
		},
		participants: [
			{
				id: 'alex',
				profileId: 'alex-profile',
				displayNameSnapshot: 'Alex',
				gameAlias: '',
				playerNumber: 1,
				status: 'active',
				roleKey: 'scout',
				outcome: 'unset'
			},
			{
				id: 'sam',
				profileId: 'sam-profile',
				displayNameSnapshot: 'Sam',
				gameAlias: '',
				playerNumber: 2,
				status: 'eliminated',
				roleKey: 'scout',
				outcome: 'win'
			},
			{
				id: 'removed',
				profileId: 'removed-profile',
				displayNameSnapshot: 'Removed player',
				gameAlias: '',
				playerNumber: 3,
				status: 'kicked',
				roleKey: 'scout',
				outcome: 'unset'
			}
		],
		rooms: [
			{
				id: 'general',
				key: 'general',
				label: 'General',
				kind: 'general',
				playersCanPost: true,
				readable: true,
				sendable: true,
				latestMessage: null
			}
		],
		assets: [],
		awards: [],
		audit: [],
		attentionItems: [],
		abilityProgress: {
			phaseKey: 'discussion',
			roundNumber: 3,
			eligiblePlayerCount: 1,
			activatedPlayerCount: 1,
			finalizedPlayerCount: 0,
			locked: false
		},
		abilityResults: []
	};
}

async function mockGame(page: Page, view: AdminGameView) {
	await page.addInitScript(() => {
		const token = `test.${btoa(JSON.stringify({ exp: 4_000_000_000 }))}.test`;
		localStorage.setItem(
			'pocketbase_auth',
			JSON.stringify({
				token,
				record: { id: 'actor', type: 'game_masters', displayName: 'Actions tester' }
			})
		);
	});
	await page.route('**/api/realtime', (route) =>
		route.request().method() === 'GET'
			? route.fulfill({
					contentType: 'text/event-stream',
					body: 'id: actions\nevent: PB_CONNECT\ndata: {}\n\n'
				})
			: route.fulfill({ json: {} })
	);
	const commands: Array<{ path: string; method: string; body: unknown }> = [];
	await page.route('**/api/app/**', (route) => {
		const request = route.request();
		const path = new URL(request.url()).pathname;
		if (path.endsWith('/admin-view')) return route.fulfill({ json: view });
		if (path.endsWith('/unread-counts')) return route.fulfill({ json: { counts: {}, total: 0 } });
		if (path.endsWith('/profile-requests')) return route.fulfill({ json: [] });
		if (request.method() !== 'GET')
			commands.push({ path, method: request.method(), body: request.postDataJSON() });
		return route.fulfill({
			json: { version: 'test', items: [], pendingCount: 0, recipientTotal: 1 }
		});
	});
	await page.goto('/admin/games/actionstest/overview');
	await expect(page.getByRole('link', { name: 'Game', exact: true })).toBeVisible();
	return commands;
}

async function openPlayers(page: Page) {
	await page.getByRole('link', { name: 'Players & roles', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/games\/actionstest\/players$/);
	await expect(page.getByRole('link', { name: 'Players & roles', exact: true })).toHaveAttribute(
		'aria-current',
		'page'
	);
}

function playerRow(page: Page, player: string) {
	return page
		.getByRole('list', { name: 'Players', exact: true })
		.getByRole('listitem')
		.filter({ hasText: player === 'alex' ? 'Alex' : 'Sam' });
}

async function playerAction(page: Page, player: string, action: string) {
	await openPlayers(page);
	await playerRow(page, player).getByRole('button', { name: 'Actions', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Player actions', exact: true });
	await expect(
		dialog.getByText(player === 'alex' ? 'Player 1 · Alex' : 'Player 2 · Sam')
	).toBeVisible();
	await expect(dialog.getByRole('combobox', { name: 'Player', exact: true })).toHaveCount(0);
	await dialog.getByRole('combobox', { name: 'Action', exact: true }).selectOption(action);
	return dialog;
}

async function openAnnouncements(page: Page) {
	await expect(page.getByRole('complementary', { name: 'Announcements' })).toHaveCount(0);
	await expect(page.getByRole('button', { name: 'Send announcement', exact: true })).toHaveCount(0);
	const navigation = page.getByRole('navigation', { name: 'Live game', exact: true });
	await navigation.getByRole('link', { name: 'Announcements', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/games\/actionstest\/announcements$/);
	await expect(
		navigation.getByRole('link', { name: 'Announcements', exact: true })
	).toHaveAttribute('aria-current', 'page');
	await expect(page.getByRole('navigation', { name: 'Management', exact: true })).toHaveCount(0);
	await expect(page.getByRole('heading', { name: 'Announcements', level: 1 })).toBeVisible();
}

for (const width of [320, 1280]) {
	test(`actions and focused sheets work at ${width}px`, async ({ page }, testInfo) => {
		await page.setViewportSize({ width, height: 900 });
		if (width === 320)
			await page.addInitScript(() =>
				localStorage.setItem(
					'sgh.display-preferences.v1',
					JSON.stringify({ largeText: true, highContrast: true })
				)
			);
		const view = gameView();
		view.game.rolesVisible = false;
		const commands = await mockGame(page, view);
		await expect(page.getByText('Roles are hidden from players', { exact: true })).toHaveCount(0);
		await expect(page.getByRole('button', { name: 'Reveal roles', exact: true })).toHaveCount(0);
		await expect(page.getByRole('button', { name: 'Pause game', exact: true })).toHaveCount(0);
		await expect(page.getByRole('button', { name: 'Resume timer', exact: true })).toBeVisible();
		await page.getByRole('button', { name: 'Change phase', exact: true }).click();
		const phase = page.getByRole('dialog', { name: 'Change phase', exact: true });
		await expect(phase.getByRole('combobox', { name: 'Phase', exact: true })).toHaveValue('voting');
		await expect(phase.getByText('Starts round 4.')).toBeVisible();
		await expect(phase.getByText('Current ability choices will be finalized.')).toBeVisible();
		await phase.getByRole('checkbox', { name: /Start the suggested timer/ }).uncheck();
		await phase.getByRole('button', { name: 'Start Voting', exact: true }).click();
		await expect(phase).not.toBeVisible();
		expect(commands).toContainEqual({
			path: '/api/app/v1/games/actionstest/phase',
			method: 'POST',
			body: { phaseKey: 'voting' }
		});
		expect(commands.some((command) => command.path.endsWith('/timer/start'))).toBe(false);
		await expect(page.getByRole('button', { name: 'Player actions', exact: true })).toHaveCount(0);
		await openPlayers(page);
		await expect(
			page
				.getByRole('listitem')
				.filter({ hasText: 'Removed player' })
				.getByRole('button', { name: 'Actions', exact: true })
		).toHaveCount(0);
		const eliminate = playerRow(page, 'alex').getByRole('button', { name: 'Actions', exact: true });
		await eliminate.click();
		const dialog = page.getByRole('dialog', { name: 'Player actions', exact: true });
		await expect(dialog.getByText('Player 1 · Alex', { exact: true })).toBeVisible();
		const action = dialog.getByRole('combobox', { name: 'Action', exact: true });
		await expect(action.getByRole('option', { name: 'Reinstate player', exact: true })).toHaveCount(
			0
		);
		await action.selectOption('eliminate');
		await expect(dialog.getByRole('button', { name: 'Eliminate Alex', exact: true })).toBeEnabled();
		await page.keyboard.press('Escape');
		await expect(dialog).not.toBeVisible();
		await expect(eliminate).toBeFocused();
		await page.screenshot({ path: testInfo.outputPath(`players-${width}.png`), fullPage: true });
		expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(
			true
		);
	});
}

for (const status of ['lobby', 'paused', 'review', 'archived'] as const) {
	test(`player actions respect ${status} game availability`, async ({ page }) => {
		const view = gameView(status);
		view.ruleset.achievements = [];
		await mockGame(page, view);
		await openPlayers(page);
		const alex = playerRow(page, 'alex').getByRole('button', { name: 'Actions', exact: true });
		if (status === 'lobby' || status === 'archived') {
			await expect(alex).toHaveCount(0);
		} else {
			await alex.click();
			const dialog = page.getByRole('dialog', { name: 'Player actions', exact: true });
			await expect(dialog.getByRole('combobox', { name: 'Action' }).getByRole('option')).toHaveText(
				status === 'paused'
					? ['Choose an action', 'Eliminate player', 'Set outcome']
					: ['Choose an action', 'Set outcome']
			);
			await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
			await expect(dialog).not.toBeVisible();
		}
		const sam = playerRow(page, 'sam').getByRole('button', { name: 'Actions', exact: true });
		if (status === 'archived') {
			await expect(sam).toHaveCount(0);
		} else {
			await sam.click();
			const dialog = page.getByRole('dialog', { name: 'Player actions', exact: true });
			await expect(dialog.getByRole('combobox', { name: 'Action' }).getByRole('option')).toHaveText(
				status === 'lobby'
					? ['Choose an action', 'Reinstate player']
					: ['Choose an action', 'Reinstate player', 'Set outcome']
			);
		}
	});
}

test('player action failures keep the target and allow a retry', async ({ page }) => {
	const view = gameView();
	await mockGame(page, view);
	let failed = false;
	await page.route('**/games/actionstest/participants/alex/eliminate', (route) => {
		if (!failed) {
			failed = true;
			return route.fulfill({
				status: 409,
				json: { code: 'test.conflict', message: 'The game changed. Try again.' }
			});
		}
		view.participants[0].status = 'eliminated';
		return route.fulfill({ json: view.participants[0] });
	});
	const dialog = await playerAction(page, 'alex', 'eliminate');
	await dialog.getByRole('button', { name: 'Eliminate Alex', exact: true }).click();
	await expect(dialog.getByRole('alert')).toContainText('The game changed. Try again.');
	await expect(dialog.getByText('Player 1 · Alex', { exact: true })).toBeVisible();
	await dialog.getByRole('button', { name: 'Eliminate Alex', exact: true }).click();
	await expect(dialog).not.toBeVisible();
	await expect(page.getByText('Alex eliminated.', { exact: true })).toBeVisible();
	const updated = await playerAction(page, 'alex', 'reinstate');
	await expect(updated.getByText('Eliminated', { exact: false })).toBeVisible();
	await expect(
		updated
			.getByRole('combobox', { name: 'Action' })
			.getByRole('option', { name: 'Eliminate player', exact: true })
	).toHaveCount(0);
});

test('outcomes start with the saved value and achievements include the selected player and note', async ({
	page
}) => {
	const commands = await mockGame(page, gameView());
	let dialog = await playerAction(page, 'sam', 'outcome');
	await expect(dialog.getByRole('combobox', { name: 'Outcome', exact: true })).toHaveValue('win');
	await dialog.getByRole('combobox', { name: 'Outcome', exact: true }).selectOption('draw');
	await dialog.getByRole('button', { name: 'Save outcome', exact: true }).click();
	await expect(dialog).not.toBeVisible();
	expect(commands).toContainEqual({
		path: '/api/app/v1/games/actionstest/outcomes',
		method: 'PUT',
		body: { outcomes: [{ participantId: 'sam', outcome: 'draw' }] }
	});
	dialog = await playerAction(page, 'alex', 'achievement');
	await dialog.getByRole('combobox', { name: 'Achievement', exact: true }).selectOption('hero');
	await dialog.getByRole('textbox', { name: 'Note (optional)' }).fill('Helped Sam');
	await dialog.getByRole('button', { name: 'Award achievement', exact: true }).click();
	await expect(dialog).not.toBeVisible();
	expect(commands).toContainEqual({
		path: '/api/app/v1/games/actionstest/achievement-awards',
		method: 'POST',
		body: { participantId: 'alex', achievementId: 'hero', note: 'Helped Sam' }
	});
});

test('announcements name the audience, count current participants and preserve drafts after errors', async ({
	page
}) => {
	const commands = await mockGame(page, gameView());
	await openAnnouncements(page);
	await page.getByRole('button', { name: 'Send announcement', exact: true }).click();
	const dialog = page.getByRole('dialog', { name: 'Send announcement', exact: true });
	await expect(dialog.getByRole('button', { name: 'Send to everyone' })).toBeDisabled();
	await dialog.getByRole('textbox', { name: 'Announcement message' }).fill('Gather at the table');
	await dialog.getByRole('combobox', { name: 'Recipients', exact: true }).selectOption('team');
	await dialog.getByRole('combobox', { name: 'Team', exact: true }).selectOption('blue');
	await expect(dialog.getByRole('status')).toHaveText('2 players will receive this announcement.');
	await expect(dialog.getByRole('button', { name: 'Send to Blue Team' })).toBeEnabled();
	await dialog.getByRole('combobox', { name: 'Recipients', exact: true }).selectOption('player');
	await dialog.getByRole('combobox', { name: 'Player', exact: true }).selectOption('alex');
	await page.route('**/api/app/v1/games/actionstest/announcements', (route) =>
		route.fulfill({
			status: 422,
			json: { code: 'test.invalid', message: 'Please shorten the message.' }
		})
	);
	await dialog.getByRole('button', { name: 'Send to Alex' }).click();
	await expect(dialog.getByRole('alert')).toContainText('Please shorten the message.');
	await expect(dialog.getByRole('textbox', { name: 'Announcement message' })).toHaveValue(
		'Gather at the table'
	);
	await page.unroute('**/api/app/v1/games/actionstest/announcements');
	await dialog.getByRole('button', { name: 'Send to Alex' }).click();
	await expect(dialog).not.toBeVisible();
	expect(commands).toContainEqual({
		path: '/api/app/v1/games/actionstest/announcements',
		method: 'POST',
		body: {
			content: 'Gather at the table',
			audience: 'player',
			targetId: 'alex',
			cueKey: '',
			imageAssetKey: '',
			imageDescription: '',
			audioAssetKey: '',
			audioAlternative: ''
		}
	});
	await expect(page.getByRole('button', { name: 'Send announcement', exact: true })).toBeFocused();
});

for (const width of [320, 1280]) {
	test(`announcement panel stays stable through sending, acknowledgements and clearing at ${width}px`, async ({
		page
	}) => {
		await page.setViewportSize({ width, height: 900 });
		if (width === 320)
			await page.addInitScript(() =>
				localStorage.setItem(
					'sgh.display-preferences.v1',
					JSON.stringify({ largeText: true, highContrast: true })
				)
			);
		const view = gameView();
		const commands = await mockGame(page, view);
		await openAnnouncements(page);
		const panel = page.getByRole('complementary', { name: 'Announcements', exact: true });
		await expect(panel.getByRole('heading', { name: 'Announcements', exact: true })).toBeVisible();
		await page.evaluate(() => document.fonts.ready);
		const initial = await panel.boundingBox();
		await expect(panel.getByRole('button')).toHaveCount(1);
		await expect(page.getByRole('button', { name: 'View acknowledgements' })).toHaveCount(0);
		const announcement: AdminAttentionSummary = {
			id: 'notice',
			kind: 'announcement',
			content: 'Gather at the table',
			senderLabel: 'Host',
			createdAt: new Date().toISOString(),
			audience: 'all',
			recipientTotal: 30,
			acknowledgementCount: 0,
			unacknowledgedPlayers: Array.from({ length: 30 }, (_, index) => ({
				participantId: `player-${index}`,
				playerNumber: index + 1,
				displayName: `Player with a long name ${index + 1}`
			}))
		};
		await page.route('**/api/app/v1/games/actionstest/announcements', (route) => {
			view.attentionItems = [announcement];
			return route.fulfill({ json: announcement });
		});
		await panel.getByRole('button', { name: 'Send announcement', exact: true }).click();
		const dialog = page.getByRole('dialog', { name: 'Send announcement', exact: true });
		await dialog.getByRole('textbox', { name: 'Announcement message' }).fill(announcement.content);
		await dialog.getByRole('button', { name: 'Send to everyone', exact: true }).click();
		await expect(dialog).not.toBeVisible();
		await expect(panel.getByRole('heading', { name: 'Not acknowledged (30)' })).toBeVisible();
		await expect(panel.getByRole('listitem')).toHaveCount(30);
		const expectStable = async () => {
			const box = await panel.boundingBox();
			expect(box!.width).toBe(initial!.width);
			expect(box!.height).toBe(initial!.height);
			expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(
				true
			);
		};
		await expectStable();
		await panel.getByRole('button', { name: 'Send reminder', exact: true }).click();
		await expect(page.getByText('Reminder sent.', { exact: true })).toBeVisible();
		expect(commands).toContainEqual({
			path: '/api/app/v1/games/actionstest/announcements/notice/remind',
			method: 'POST',
			body: {}
		});
		announcement.unacknowledgedPlayers = [announcement.unacknowledgedPlayers![1]];
		announcement.acknowledgementCount = 29;
		await page.reload();
		await expect(panel.getByRole('listitem')).toHaveCount(1);
		await expect(panel.getByText('Player 2 · Player with a long name 2')).toBeVisible();
		await expectStable();
		announcement.unacknowledgedPlayers = [];
		announcement.acknowledgementCount = 30;
		await page.reload();
		await expect(panel.getByText('All players acknowledged.', { exact: true })).toBeVisible();
		await expect(panel.getByRole('button', { name: 'Send reminder', exact: true })).toBeDisabled();
		await expectStable();
		await page.route('**/announcements/notice/clear', (route) => {
			announcement.trackingCleared = true;
			return route.fulfill({ status: 204 });
		});
		await panel.getByRole('button', { name: 'Clear', exact: true }).click();
		await expect(panel.getByRole('button')).toHaveCount(1);
		await expect(
			panel.getByRole('button', { name: 'Send announcement', exact: true })
		).toBeFocused();
		await expect(panel.getByText(announcement.content, { exact: true })).toHaveCount(0);
		await expectStable();
		await page.reload();
		await expect(panel.getByRole('button')).toHaveCount(1);
		await expectStable();
	});
}

test('announcement reminder failures keep pending players and allow retry', async ({ page }) => {
	const view = gameView();
	view.attentionItems = [
		{
			id: 'notice',
			kind: 'announcement',
			content: 'Gather',
			senderLabel: 'Host',
			createdAt: new Date().toISOString(),
			audience: 'all',
			recipientTotal: 2,
			acknowledgementCount: 1,
			unacknowledgedPlayers: [{ participantId: 'sam', playerNumber: 2, displayName: 'Sam' }]
		}
	];
	await mockGame(page, view);
	await openAnnouncements(page);
	let attempts = 0;
	await page.route('**/announcements/notice/remind', (route) =>
		++attempts === 1
			? route.fulfill({
					status: 409,
					json: { code: 'test.conflict', message: 'Please try again.' }
				})
			: route.fulfill({ json: view.attentionItems[0] })
	);
	const panel = page.getByRole('complementary', { name: 'Announcements' });
	await panel.getByRole('button', { name: 'Send reminder', exact: true }).click();
	await expect(panel.getByRole('alert')).toContainText('Please try again.');
	await expect(panel.getByRole('listitem')).toHaveText('Player 2 · Sam');
	await panel.getByRole('button', { name: 'Send reminder', exact: true }).click();
	await expect(page.getByText('Reminder sent.', { exact: true })).toBeVisible();
	await expect(panel.getByRole('alert')).toHaveCount(0);
});

for (const width of [320, 1280]) {
	test(`chat locks and unlocks all channels at ${width}px`, async ({ page }, testInfo) => {
		await page.setViewportSize({ width, height: 900 });
		const view = gameView();
		view.rooms.push(
			{
				...view.rooms[0],
				id: 'team',
				key: 'team:blue',
				kind: 'team',
				label: 'Blue Team',
				playersCanPost: false
			},
			{ ...view.rooms[0], id: 'custom', key: 'custom:reports', kind: 'custom', label: 'Reports' },
			{ ...view.rooms[0], id: 'direct', key: 'gm:alex', kind: 'gm_dm', label: 'Alex' },
			{
				...view.rooms[0],
				id: 'private',
				key: 'dm:alex:sam',
				kind: 'player_dm',
				label: 'Alex and Sam'
			}
		);
		await mockGame(page, view);
		await expect(
			page.getByRole('button', { name: 'Manage player posting', exact: true })
		).toHaveCount(0);
		await expect(page.getByRole('button', { name: 'Lock all channels', exact: true })).toHaveCount(
			0
		);
		const changes: string[] = [];
		await page.route('**/api/app/v1/games/actionstest/rooms', (route) =>
			route.fulfill({ json: view.rooms })
		);
		await page.route('**/api/app/v1/rooms/*/messages', (route) =>
			route.fulfill({ json: { items: [], nextCursor: '' } })
		);
		await page.route('**/api/app/v1/rooms/*', (route) => {
			expect(route.request().method()).toBe('PATCH');
			const room = view.rooms.find(
				(room) => room.id === new URL(route.request().url()).pathname.split('/').at(-1)
			)!;
			changes.push(room.id);
			room.playersCanPost = route.request().postDataJSON().playersCanPost;
			return route.fulfill({ json: room });
		});
		await page.getByRole('link', { name: 'Chat', exact: true }).click();
		const lock = page.getByRole('button', { name: 'Lock all channels', exact: true });
		await expect(lock).toBeVisible();
		const rail = page.getByRole('heading', { name: 'Conversations', exact: true });
		expect((await lock.boundingBox())!.y + (await lock.boundingBox())!.height).toBeLessThan(
			(await rail.boundingBox())!.y
		);
		await lock.click();
		const unlock = page.getByRole('button', { name: 'Unlock all channels', exact: true });
		await expect(unlock).toBeEnabled();
		expect(changes.toSorted()).toEqual(['custom', 'direct', 'general', 'private']);
		for (const label of ['General', 'Blue Team', 'Reports', 'Alex', 'Alex and Sam']) {
			await expect(
				page.getByRole('button', { name: `${label}, Players read-only`, exact: true })
			).toBeEnabled();
		}
		await expect(
			page.getByText('Players cannot post. They can still read messages.', { exact: true })
		).toBeVisible();
		await page.getByRole('button', { name: 'Dismiss notification', exact: true }).click();
		await page.screenshot({
			path: testInfo.outputPath(`locked-chat-${width}.png`),
			fullPage: true
		});
		await page.getByRole('button', { name: 'General, Players read-only', exact: true }).click();
		await expect(page.getByRole('heading', { name: 'General', exact: true })).toBeVisible();
		await expect(page.getByRole('textbox', { name: 'Message', exact: true })).toBeVisible();
		if (width === 320) {
			const composer = await page
				.getByRole('textbox', { name: 'Message', exact: true })
				.boundingBox();
			const navigation = await page.getByRole('link', { name: 'Chat', exact: true }).boundingBox();
			expect(composer!.y + composer!.height).toBeLessThanOrEqual(navigation!.y);
		}
		await unlock.click();
		await expect(lock).toBeEnabled();
		expect(changes.toSorted()).toEqual([
			'custom',
			'custom',
			'direct',
			'direct',
			'general',
			'general',
			'private',
			'private',
			'team'
		]);
		if (width === 320) await page.getByRole('button', { name: 'Back to conversations' }).click();
		for (const label of ['Alex', 'Alex and Sam']) {
			await expect(page.getByRole('button', { name: label, exact: true })).toBeEnabled();
		}
		if (width === 320) await page.getByRole('button', { name: 'General', exact: true }).click();
		await page.getByRole('button', { name: 'Dismiss notification', exact: true }).click();
		await expect(page.getByRole('button', { name: 'Lock channel', exact: true })).toBeEnabled();
		const unlockedBounds = await page
			.getByRole('button', { name: 'Lock channel', exact: true })
			.boundingBox();
		await page.getByRole('button', { name: 'Lock channel', exact: true }).click();
		await expect(page.getByText('Players read-only', { exact: false })).toBeVisible();
		await expect(page.getByRole('button', { name: 'Unlock channel', exact: true })).toBeEnabled();
		await page.getByRole('heading', { name: 'General', exact: true }).hover();
		await expect
			.poll(() => page.getByRole('button', { name: 'Unlock channel', exact: true }).boundingBox())
			.toEqual(unlockedBounds);
		await page.getByRole('button', { name: 'Dismiss notification', exact: true }).click();
		await page.screenshot({
			path: testInfo.outputPath(`channel-header-${width}.png`),
			fullPage: true
		});
		await page.getByRole('button', { name: 'Unlock channel', exact: true }).click();
		await expect(page.getByRole('button', { name: 'Lock channel', exact: true })).toBeEnabled();
		await page.getByRole('heading', { name: 'General', exact: true }).hover();
		await expect
			.poll(() => page.getByRole('button', { name: 'Lock channel', exact: true }).boundingBox())
			.toEqual(unlockedBounds);
		await expect(page.getByText('Players read-only', { exact: false })).not.toBeVisible();
	});
}

test('rulesets without phases retain live controls without showing phase controls', async ({
	page
}) => {
	const view = gameView();
	view.ruleset.phases = [];
	view.game.phaseKey = '';
	view.abilityProgress.eligiblePlayerCount = 0;
	await mockGame(page, view);
	await expect(page.getByRole('button', { name: 'Change phase', exact: true })).toHaveCount(0);
	await expect(page.getByText('Open play', { exact: true })).toHaveCount(0);
	await expect(page.getByRole('region', { name: 'Game timer', exact: true })).toBeVisible();
	await expect(page.getByRole('button', { name: 'Pause game', exact: true })).toHaveCount(0);
	await expect(page.getByRole('button', { name: 'Player actions', exact: true })).toHaveCount(0);
	const dialog = await playerAction(page, 'alex', 'eliminate');
	await expect(dialog.getByRole('button', { name: 'Eliminate Alex', exact: true })).toBeEnabled();
});

test('the row action form keeps its player while changing actions and resets for another player', async ({
	page
}) => {
	await mockGame(page, gameView());
	const dialog = await playerAction(page, 'sam', 'outcome');
	await expect(dialog.getByText('Sam · Eliminated', { exact: true })).toBeVisible();
	const action = dialog.getByRole('combobox', { name: 'Action', exact: true });
	await action.selectOption('outcome');
	await expect(dialog.getByRole('combobox', { name: 'Outcome' })).toHaveValue('win');
	await action.selectOption('achievement');
	await expect(dialog.getByText('Player 2 · Sam', { exact: true })).toBeVisible();
	await expect(dialog.getByRole('combobox', { name: 'Achievement' })).toBeVisible();
	await dialog.getByRole('combobox', { name: 'Achievement' }).selectOption('hero');
	await dialog.getByRole('textbox', { name: 'Note (optional)' }).fill('A note for Sam');
	await dialog.getByRole('button', { name: 'Cancel', exact: true }).click();
	await expect(dialog).not.toBeVisible();
	await playerRow(page, 'alex').getByRole('button', { name: 'Actions', exact: true }).click();
	await expect(dialog.getByText('Player 1 · Alex', { exact: true })).toBeVisible();
	await expect(action).toHaveValue('');
	await action.selectOption('achievement');
	await expect(dialog.getByRole('combobox', { name: 'Achievement' })).toHaveValue('');
	await dialog.getByRole('combobox', { name: 'Achievement' }).selectOption('hero');
	await expect(dialog.getByRole('textbox', { name: 'Note (optional)' })).toHaveValue('');
});

test('reinstate targets eliminated players and an existing achievement can be revoked', async ({
	page
}) => {
	const view = gameView();
	view.awards.push({
		id: 'award-1',
		profileId: 'sam-profile',
		achievementId: 'hero',
		title: 'Hero',
		description: 'Help another player.',
		points: 1,
		hiddenUntilGameCompleted: false,
		awardedAt: new Date().toISOString()
	});
	const commands = await mockGame(page, view);
	let dialog = await playerAction(page, 'sam', 'reinstate');
	await expect(dialog.getByText('Player 2 · Sam', { exact: true })).toBeVisible();
	await expect(
		dialog
			.getByRole('combobox', { name: 'Action' })
			.getByRole('option', { name: 'Eliminate player', exact: true })
	).toHaveCount(0);
	await dialog.getByRole('button', { name: 'Reinstate Sam', exact: true }).click();
	await expect(dialog).not.toBeVisible();
	expect(commands).toContainEqual({
		path: '/api/app/v1/games/actionstest/participants/sam/reinstate',
		method: 'POST',
		body: {}
	});
	dialog = await playerAction(page, 'sam', 'achievement');
	await dialog.getByRole('combobox', { name: 'Achievement', exact: true }).selectOption('hero');
	await expect(
		dialog.getByText('Sam already has this achievement. You can revoke it.')
	).toBeVisible();
	await dialog.getByRole('button', { name: 'Revoke achievement', exact: true }).click();
	await expect(dialog).not.toBeVisible();
	expect(
		commands.some(
			(command) =>
				command.path.endsWith('/achievement-awards/award-1') && command.method === 'DELETE'
		)
	).toBe(true);
});

test('finishing enters review through the existing completion endpoint', async ({ page }) => {
	const view = gameView();
	await mockGame(page, view);
	let finished = false;
	await page.route('**/games/actionstest/completion/start', (route) => {
		finished = route.request().method() === 'POST';
		view.game.status = 'review';
		return route.fulfill({ json: view.game });
	});
	await page.getByRole('button', { name: 'Finish game', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/games\/actionstest\/finish\/outcomes$/);
	expect(finished).toBe(true);
});

test('lobby readiness prevents starting without valid roles', async ({ page }) => {
	const view = gameView('lobby');
	view.participants[0].roleKey = '';
	await mockGame(page, view);
	await expect(page.getByRole('button', { name: 'Start game', exact: true })).toBeDisabled();
	await expect(page.getByRole('link', { name: 'Assign roles', exact: true })).toBeVisible();
	await expect(
		page.getByText('Assign a role to the remaining player before starting.')
	).toBeVisible();
	await expect(page.getByRole('region', { name: 'Game timer', exact: true })).toHaveCount(0);
	await expect(page.getByRole('button', { name: 'Eliminate player', exact: true })).toHaveCount(0);
	view.participants[0].roleKey = 'scout';
	await page.reload();
	await expect(page.getByRole('button', { name: 'Start game', exact: true })).toBeEnabled();
});

for (const status of ['review', 'archived'] as const) {
	test(`${status} games show their appropriate destination`, async ({ page }) => {
		await mockGame(page, gameView(status));
		await expect(
			page.getByRole('link', {
				name: status === 'review' ? 'Continue finishing' : 'View summary',
				exact: true
			})
		).toHaveAttribute(
			'href',
			`/admin/games/actionstest/${status === 'review' ? 'finish/outcomes' : 'summary'}`
		);
		await expect(page.getByRole('button', { name: 'Send announcement', exact: true })).toHaveCount(
			0
		);
		await expect(page.getByRole('region', { name: 'Game timer', exact: true })).toHaveCount(0);
		await openAnnouncements(page);
		await expect(page.getByRole('button', { name: 'Send announcement', exact: true })).toHaveCount(
			0
		);
		await expect(
			page.getByText('Announcements can be sent while the game is in the lobby, running or paused.')
		).toBeVisible();
	});
}

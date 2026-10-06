import { type Page, type Locator } from '@playwright/test';
import { expect, test } from './fixtures';

const game = {
	id: 'clocktest',
	name: 'Clock test',
	status: 'running',
	phaseKey: 'discussion',
	revision: 1
};
const timer = (status: string) => ({
	status,
	remainingMs: status === 'completed' ? 0 : 65_000,
	totalMs: 65_000,
	revision: 1,
	serverTime: '2026-01-01T00:00:00.000Z'
});

async function mockGame(page: Page, admin = false, timerState = { status: 'inactive' }) {
	await page.addInitScript(
		(type) => {
			const token = `test.${btoa(JSON.stringify({ exp: 4_000_000_000 }))}.test`;
			localStorage.setItem(
				'pocketbase_auth',
				JSON.stringify({ token, record: { id: 'actor', type, displayName: 'Clock tester' } })
			);
		},
		admin ? 'game_masters' : 'player_profiles'
	);
	await page.route('**/api/realtime', (route) =>
		route.request().method() === 'GET'
			? route.fulfill({
					contentType: 'text/event-stream',
					body: 'id: clocktest\nevent: PB_CONNECT\ndata: {}\n\n'
				})
			: route.fulfill({ json: {} })
	);
	await page.route('**/api/app/**', (route) => {
		const path = new URL(route.request().url()).pathname;
		if (path.endsWith('/player-view'))
			return route.fulfill({
				json: {
					game,
					participant: { id: 'participant' },
					ruleset: { name: 'Test' },
					roleAvailable: false,
					rooms: [],
					attentionItems: [],
					assets: [],
					party: []
				}
			});
		if (path.endsWith('/admin-view'))
			return route.fulfill({
				json: {
					game,
					timer: timer(timerState.status),
					ruleset: { phases: [{ id: 'discussion', name: 'Discussion' }], roles: [], teams: [] },
					participants: [],
					rooms: [],
					attentionItems: [],
					assets: [],
					phaseResults: [],
					abilityProgress: { eligiblePlayerCount: 0 },
					abilityResults: [],
					awards: [],
					audit: []
				}
			});
		return route.fulfill({ json: { version: 'test', items: [], pendingCount: 0 } });
	});
}

async function dimensions(locator: Locator) {
	const box = await locator.boundingBox();
	expect(box).not.toBeNull();
	return { width: box!.width, height: box!.height };
}

async function mockTimerCommands(page: Page) {
	let revision = 1;
	const commands: { command: string | undefined; body: unknown }[] = [];
	await page.route('**/api/app/v1/games/clocktest/timer/*', (route) => {
		const command = new URL(route.request().url()).pathname.split('/').at(-1);
		commands.push({ command, body: route.request().postDataJSON() });
		return route.fulfill({
			json: {
				...timer(command === 'stop' ? 'inactive' : command === 'pause' ? 'paused' : 'running'),
				revision: ++revision
			}
		});
	});
	return commands;
}

for (const width of [320, 390, 1280, 1920]) {
	test(`player clock keeps its space while loading and across states at ${width}px`, async ({
		page
	}) => {
		await page.setViewportSize({ width, height: 800 });
		await mockGame(page);
		let state = 'inactive';
		let release!: () => void;
		let pending = new Promise<void>((resolve) => (release = resolve));
		await page.route('**/api/app/v1/games/clocktest/timer', async (route) => {
			await pending;
			await route.fulfill({ json: timer(state) });
		});
		await page.goto('/play/game');
		const placeholder = page.getByLabel('Timer inactive');
		await expect(placeholder).toHaveText('--:--');
		await page.evaluate(() => document.fonts.ready);
		const clock = placeholder.locator('..');
		const size = await dimensions(clock);
		const party = page.getByRole('link', { name: /View party/ });
		const partyPosition = await party.boundingBox();
		release();
		for (state of ['running', 'paused', 'completed', 'inactive']) {
			let nextRelease!: () => void;
			pending = new Promise<void>((resolve) => (nextRelease = resolve));
			await page.reload();
			await expect(page.getByLabel('Timer inactive')).toHaveText('--:--');
			nextRelease();
			const time =
				state === 'inactive'
					? page.getByLabel('Timer inactive')
					: page.getByLabel(`${state === 'completed' ? 0 : 65} seconds remaining`);
			await expect(time).toHaveText(
				state === 'inactive' ? '--:--' : state === 'completed' ? '00:00' : '01:05'
			);
			expect(await dimensions(time.locator('..'))).toEqual(size);
			expect(await party.boundingBox()).toEqual(partyPosition);
		}
	});

	test(`admin clock keeps its space across states at ${width}px`, async ({ page }) => {
		await page.setViewportSize({ width, height: 800 });
		const timerState = { status: 'inactive' };
		await mockGame(page, true, timerState);
		const commands = await mockTimerCommands(page);
		await page.goto('/admin/games/clocktest/overview');
		await expect(page.getByLabel('Timer inactive')).toHaveText('--:--');
		await page.evaluate(() => document.fonts.ready);
		const timerRegion = page.getByRole('region', { name: 'Game timer', exact: true });
		const card = timerRegion.locator('..').locator('..');
		const cardSize = await dimensions(card);
		expect(cardSize.width).toBeLessThanOrEqual(560);
		expect(cardSize.height).toBeLessThan(width < 768 ? 250 : 200);
		const duration = timerRegion.getByRole('combobox', { name: 'Duration', exact: true });
		const readoutBox = await page.getByLabel('Timer inactive').boundingBox();
		const durationBox = await duration.boundingBox();
		expect(readoutBox!.y + readoutBox!.height).toBeLessThan(durationBox!.y);
		const digits = await page.getByLabel('Timer inactive').evaluate((el) => {
			const range = document.createRange();
			range.selectNodeContents(el);
			const rect = range.getBoundingClientRect();
			return { x: rect.x, width: rect.width };
		});
		const initialCardBox = await card.boundingBox();
		expect(
			Math.abs(digits.x + digits.width / 2 - initialCardBox!.x - initialCardBox!.width / 2)
		).toBeLessThanOrEqual(1);
		expect(readoutBox!.height).toBeGreaterThanOrEqual(60);
		const start = timerRegion.getByRole('button', { name: 'Start timer', exact: true });
		await expect(timerRegion.getByRole('button', { name: 'Pause timer', exact: true })).toHaveCount(
			0
		);
		await duration.selectOption('3');
		const startBox = await start.boundingBox();
		expect(startBox!.height).toBe(44);
		if (width >= 1280) {
			expect(
				Math.abs(startBox!.y + startBox!.height - durationBox!.y - durationBox!.height)
			).toBeLessThanOrEqual(1);
		}
		async function expectContainedControls() {
			const bounds = await card.boundingBox();
			const controls: { x: number; width: number }[] = [];
			for (const control of await timerRegion.getByRole('button').all()) {
				const box = await control.boundingBox();
				controls.push(box!);
				expect(box!.width).toBeGreaterThanOrEqual(44);
				expect(box!.height).toBe(44);
				expect(box!.y).toBeGreaterThan(readoutBox!.y + readoutBox!.height);
				expect(box!.x).toBeGreaterThanOrEqual(bounds!.x);
				expect(box!.x + box!.width).toBeLessThanOrEqual(bounds!.x + bounds!.width);
			}
			const regionBounds = await timerRegion.boundingBox();
			if (width < 768 || controls.length > 1) {
				expect(Math.min(...controls.map((box) => box.x))).toBe(regionBounds!.x);
				expect(Math.max(...controls.map((box) => box.x + box.width))).toBeCloseTo(
					regionBounds!.x + regionBounds!.width,
					1
				);
			}
			expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(
				true
			);
		}
		await expectContainedControls();
		const size = await dimensions(page.getByLabel('Timer inactive').locator('..'));
		const cardBox = await card.boundingBox();
		const gamePauseBox = await page
			.getByRole('button', { name: 'Pause game', exact: true })
			.boundingBox();
		for (const name of ['Start timer', 'Pause timer', 'Resume timer', 'Clear']) {
			await page.getByRole('button', { name, exact: true }).click();
			await page.mouse.move(0, 0);
			const time =
				name === 'Clear'
					? page.getByLabel('Timer inactive')
					: page.getByLabel('65 seconds remaining');
			await expect(time).toHaveText(name === 'Clear' ? '--:--' : '01:05');
			expect(await dimensions(time.locator('..'))).toEqual(size);
			expect(await card.boundingBox()).toEqual(cardBox);
			expect(await time.boundingBox()).toEqual(readoutBox);
			const nextPrimary = timerRegion.getByRole('button', {
				name:
					name === 'Pause timer'
						? 'Resume timer'
						: name === 'Clear'
							? 'Start timer'
							: 'Pause timer',
				exact: true
			});
			await expect.poll(() => nextPrimary.boundingBox()).toEqual(startBox);
			if (name !== 'Clear') await expect(nextPrimary).toBeFocused();
			await expectContainedControls();
			await expect(page.getByRole('button', { name: 'Pause game', exact: true })).toBeVisible();
			expect(
				await page.getByRole('button', { name: 'Pause game', exact: true }).boundingBox()
			).toEqual(gamePauseBox);
		}
		expect(commands).toEqual([
			{ command: 'start', body: { durationMs: 180_000 } },
			{ command: 'pause', body: {} },
			{ command: 'resume', body: {} },
			{ command: 'stop', body: {} }
		]);
		timerState.status = 'completed';
		await page.reload();
		const completed = page.getByLabel('0 seconds remaining');
		await expect(completed).toHaveText('00:00');
		expect(await dimensions(completed.locator('..'))).toEqual(size);
		expect(await card.boundingBox()).toEqual(cardBox);
		expect(
			await timerRegion.getByRole('button', { name: 'Start again', exact: true }).boundingBox()
		).toEqual(startBox);
	});
}

test('admin timer reflows without overlapping at 320px with large text', async ({ page }) => {
	await page.setViewportSize({ width: 320, height: 800 });
	await mockGame(page, true);
	await mockTimerCommands(page);
	await page.addInitScript(() => {
		localStorage.setItem(
			'sgh.display-preferences.v1',
			JSON.stringify({ largeText: true, highContrast: true })
		);
	});
	await page.goto('/admin/games/clocktest/overview');
	const time = page.getByLabel('Timer inactive');
	await expect(time).toHaveText('--:--');
	await page.evaluate(() => document.fonts.ready);
	const duration = page.getByRole('combobox', { name: 'Duration', exact: true });
	const readoutBox = await time.boundingBox();
	const durationBox = await duration.boundingBox();
	const start = page.getByRole('button', { name: 'Start timer', exact: true });
	const startBox = await start.boundingBox();
	await expect(start).toBeEnabled();
	await expect(page.getByRole('button', { name: 'Pause timer', exact: true })).toHaveCount(0);
	expect(readoutBox!.y + readoutBox!.height).toBeLessThan(durationBox!.y);
	expect(readoutBox!.y + readoutBox!.height).toBeLessThan(startBox!.y);
	expect(startBox!.y + startBox!.height).toBeLessThan(durationBox!.y);
	expect(startBox!.height).toBeLessThanOrEqual(48);
	expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true);
	const region = page.getByRole('region', { name: 'Game timer', exact: true });
	const card = region.locator('..').locator('..');
	const cardBox = await card.boundingBox();
	await start.focus();
	for (const name of ['Pause timer', 'Resume timer', 'Pause timer']) {
		await page.keyboard.press('Enter');
		const primary = region.getByRole('button', { name, exact: true });
		await expect(primary).toBeEnabled();
		await expect(primary).toBeFocused();
		expect(await primary.boundingBox()).toEqual(startBox);
		expect(await card.boundingBox()).toEqual(cardBox);
		expect(await region.getByLabel('65 seconds remaining').boundingBox()).toEqual(readoutBox);
		await expect(region.getByRole('combobox', { name: 'Duration', exact: true })).toHaveCount(0);
	}
});

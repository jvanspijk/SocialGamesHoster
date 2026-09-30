import { expect, test, type Page, type Locator } from '@playwright/test';

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

for (const width of [320, 1280]) {
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
		let revision = 1;
		await page.route('**/api/app/v1/games/clocktest/timer/*', (route) => {
			const command = new URL(route.request().url()).pathname.split('/').at(-1);
			return route.fulfill({
				json: {
					...timer(command === 'stop' ? 'inactive' : command === 'pause' ? 'paused' : 'running'),
					revision: ++revision
				}
			});
		});
		await page.goto('/admin/games/clocktest/overview');
		await expect(page.getByLabel('Timer inactive')).toHaveText('--:--');
		await page.evaluate(() => document.fonts.ready);
		const size = await dimensions(page.getByLabel('Timer inactive').locator('..'));
		for (const name of ['Start timer', 'Pause timer', 'Resume timer', 'Clear']) {
			await page.getByRole('button', { name, exact: true }).click();
			const time =
				name === 'Clear'
					? page.getByLabel('Timer inactive')
					: page.getByLabel('65 seconds remaining');
			await expect(time).toHaveText(name === 'Clear' ? '--:--' : '01:05');
			expect(await dimensions(time.locator('..'))).toEqual(size);
		}
		timerState.status = 'completed';
		await page.reload();
		const completed = page.getByLabel('0 seconds remaining');
		await expect(completed).toHaveText('00:00');
		expect(await dimensions(completed.locator('..'))).toEqual(size);
	});
}

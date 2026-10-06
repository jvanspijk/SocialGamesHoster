import { expect, test } from './fixtures';

for (const admin of [true, false]) {
	test(`${admin ? 'admin' : 'player'} signs out from the header`, async ({ page }) => {
		await page.addInitScript((isAdmin) => {
			const token = `test.${btoa(JSON.stringify({ exp: 4_000_000_000 }))}.test`;
			localStorage.setItem(
				'pocketbase_auth',
				JSON.stringify({
					token,
					record: {
						id: 'actor',
						type: isAdmin ? 'game_masters' : 'player_profiles',
						displayName: 'Logout tester'
					}
				})
			);
		}, admin);
		await page.route('**/api/realtime', (route) =>
			route.request().method() === 'GET'
				? route.fulfill({
						contentType: 'text/event-stream',
						body: 'id: logout-test\nevent: PB_CONNECT\ndata: {}\n\n'
					})
				: route.fulfill({ json: {} })
		);
		await page.route('**/api/app/**', (route) => {
			const path = new URL(route.request().url()).pathname;
			if (path.endsWith('/games/live') || path.endsWith('/player-view'))
				return route.fulfill({ status: 404, json: { code: 'game.no_live_game' } });
			if (path.endsWith('/profile-requests') || path.endsWith('/games'))
				return route.fulfill({ json: [] });
			return route.fulfill({ json: { version: 'test', needsOwner: false } });
		});

		await page.goto(admin ? '/admin' : '/play');
		const signOut = page.locator('header').getByRole('button', { name: 'Sign out' });
		for (const width of [1280, 320]) {
			await page.setViewportSize({ width, height: 800 });
			await expect(signOut).toBeVisible();
			const box = await signOut.boundingBox();
			expect(box!.x).toBeGreaterThanOrEqual(0);
			expect(box!.x + box!.width).toBeLessThanOrEqual(width);
		}
		if (!admin) {
			await page.getByRole('link', { name: /Settings/ }).click();
			await expect(page.getByRole('heading', { name: 'Settings', exact: true })).toBeVisible();
			await expect(page.getByRole('button', { name: 'Sign out' })).toHaveCount(1);
			await expect(page.getByRole('heading', { name: 'Account', exact: true })).toHaveCount(0);
			await page.getByRole('link', { name: 'Player home', exact: true }).click();
		}
		const logoutRequest = page.waitForRequest('**/api/app/v1/auth/logout');
		await signOut.click();
		expect((await logoutRequest).method()).toBe('POST');
		await expect(page).toHaveURL(/\/$/);
		await expect(page.getByRole('button', { name: 'Sign out' })).toHaveCount(0);
		await expect
			.poll(() => page.evaluate(() => localStorage.getItem('pocketbase_auth')))
			.toBeNull();
	});
}

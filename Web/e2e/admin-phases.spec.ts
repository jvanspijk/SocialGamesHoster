import { expect, test } from './fixtures';

for (const status of ['running', 'paused']) {
	for (const hasPhases of [false, true]) {
		test(`admin ${status} game ${hasPhases ? 'shows' : 'hides'} phase controls with ${hasPhases ? 'one phase' : 'no phases'}`, async ({
			page
		}) => {
			await page.addInitScript(() => {
				const token = `test.${btoa(JSON.stringify({ exp: 4_000_000_000 }))}.test`;
				localStorage.setItem(
					'pocketbase_auth',
					JSON.stringify({
						token,
						record: { id: 'actor', type: 'game_masters', displayName: 'Phase tester' }
					})
				);
			});
			await page.route('**/api/realtime', (route) =>
				route.request().method() === 'GET'
					? route.fulfill({
							contentType: 'text/event-stream',
							body: 'id: phasetest\nevent: PB_CONNECT\ndata: {}\n\n'
						})
					: route.fulfill({ json: {} })
			);
			await page.route('**/api/app/**', (route) => {
				if (new URL(route.request().url()).pathname.endsWith('/admin-view')) {
					return route.fulfill({
						json: {
							game: { id: 'phasetest', name: 'Phase test', status, revision: 1 },
							ruleset: {
								phases: hasPhases ? [{ id: 'discussion', name: 'Discussion' }] : [],
								roles: [],
								teams: []
							},
							participants: [],
							rooms: [],
							assets: [],
							abilityProgress: { eligiblePlayerCount: 0 },
							abilityResults: []
						}
					});
				}
				return route.fulfill({ json: { version: 'test', items: [], pendingCount: 0 } });
			});
			await page.goto('/admin/games/phasetest/overview');
			await expect(
				page.getByRole('button', { name: status === 'running' ? 'Pause game' : 'Resume game' })
			).toBeVisible();
			const changePhase = page.getByRole('button', { name: 'Change phase', exact: true });
			if (hasPhases) {
				await expect(changePhase).toBeVisible();
				await changePhase.click();
				const dialog = page.getByRole('dialog', { name: 'Change phase', exact: true });
				await expect(dialog).toBeVisible();
				await expect(dialog.getByRole('combobox', { name: 'Phase', exact: true })).toHaveValue(
					'discussion'
				);
			} else {
				await expect(changePhase).toHaveCount(0);
			}
		});
	}
}

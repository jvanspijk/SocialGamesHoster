import { expect, test } from '@playwright/test';
let token: string;
let actor: Record<string, unknown>;
let id: string;
test.beforeAll(async ({ request }) => {
	const base = '/api/app/v1';
	const status = await (await request.get(`${base}/setup/status`)).json();
	const auth = await request.post(
		`${base}${status.needsOwner ? '/setup/owner' : '/auth/game-master/login'}`,
		{
			data: {
				username: 'partyhost',
				password: 'correct-horse-battery',
				displayName: 'Party Host',
				trustedLanAcknowledged: true
			}
		}
	);
	expect(auth.ok()).toBeTruthy();
	const login = await request.post(`${base}/auth/game-master/login`, {
		data: { username: 'partyhost', password: 'correct-horse-battery' }
	});
	({ token, actor } = await login.json());
	const headers = { Authorization: token };
	const created = await request.post(`${base}/rulesets`, {
		headers,
		data: { name: 'Navigation regression', description: '', minPlayers: 2, maxPlayers: 12 }
	});
	expect(created.ok()).toBeTruthy();
	({ id } = await created.json());
	const detail = await (await request.get(`${base}/rulesets/${id}`, { headers })).json();
	const definition = detail.definition;
	definition.teams = [{ id: 'team', name: 'Team', description: '' }];
	definition.roles = [
		{
			id: 'role',
			name: 'Role',
			description: '',
			teamId: 'team',
			categoryIds: [],
			tags: [],
			abilityIds: [],
			winCondition: ''
		}
	];
	const saved = await request.post(`${base}/rulesets/${id}/save`, {
		headers,
		data: { definition }
	});
	expect(saved.ok()).toBeTruthy();
});
test.beforeEach(async ({ page }) => {
	await page.addInitScript(
		({ token, actor }) =>
			localStorage.setItem(
				'pocketbase_auth',
				JSON.stringify({ token, record: { ...actor, collectionName: 'game_masters' } })
			),
		{ token, actor }
	);
	await page.setViewportSize({ width: 1600, height: 1000 });
});
test('opening and leaving an untouched ruleset does not prompt to save', async ({ page }) => {
	await page.goto(`/admin/rulesets/${id}/edit/metadata`);
	await expect(page.getByRole('heading', { name: 'Basics', exact: true })).toBeVisible();
	await expect(page.getByRole('combobox', { name: /Ruleset cover/ })).toBeVisible();
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await expect(page.getByRole('dialog', { name: 'Leave with unsaved changes?' })).not.toBeVisible();
});
test('saved ruleset sections render in sequence without becoming dirty', async ({ page }) => {
	const errors: string[] = [];
	page.on('pageerror', (error) => errors.push(error.message));
	await page.goto(`/admin/rulesets/${id}/edit/metadata`);
	const nav = page.getByRole('navigation', { name: 'Ruleset sections' });
	for (const [label, section] of [
		['Teams', 'teams'],
		['Roles and abilities', 'roles'],
		['Basics', 'metadata']
	]) {
		await nav.getByRole('button', { name: new RegExp(`^${label}`) }).click();
		await expect(page).toHaveURL(new RegExp(`/edit/${section}$`));
		await expect(page.getByRole('heading', { name: label, exact: true }).first()).toBeVisible();
	}
	expect(errors).toEqual([]);
	await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();
	await page.getByRole('textbox', { name: /^Name/ }).fill('Changed ruleset');
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await expect(page.getByRole('dialog', { name: 'Leave with unsaved changes?' })).toBeVisible();
});

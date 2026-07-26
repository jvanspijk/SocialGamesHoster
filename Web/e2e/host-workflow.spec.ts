import { expect, test } from '@playwright/test';

test('owner setup opens the visual ruleset creator', async ({ page }) => {
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Prepare the game table' })).toBeVisible();

	await page.getByLabel('Username').fill('partyhost');
	await page.getByLabel('Display name').fill('Party Host');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	const liveSubscription = page.waitForResponse(
		(response) => response.url().endsWith('/api/realtime') && response.request().method() === 'POST'
	);
	await page.getByRole('button', { name: 'Create owner' }).click();
	expect((await liveSubscription).status()).toBe(204);

	await expect(page.getByRole('button', { name: /rulesets/i })).toBeVisible();
	await page.getByRole('button', { name: /rulesets/i }).click();
	await page.getByRole('link', { name: /new ruleset/i }).click();

	await expect(page.getByRole('heading', { name: 'Overview and limits' })).toBeVisible();
	await page.getByLabel('Stable slug').fill('party-test');
	await page.getByLabel('Name').fill('Party Test');
	await page.getByRole('button', { name: 'Achievements' }).click();
	await page.getByRole('button', { name: 'Add achievement' }).click();
	await page.getByLabel('Achievement points').fill('75');
	await page.getByLabel('Hide from players until the game ends').check();

	await expect(page.getByLabel('Achievement points')).toHaveValue('75');
	await expect(page.getByText('Advanced JSON')).toBeVisible();
	await expect(page.getByLabel('achievements JSON')).not.toBeVisible();

	await page.getByRole('button', { name: 'Save' }).click();
	await expect(page).not.toHaveURL(/\/new$/);
	await page.reload();
	await page.getByRole('button', { name: 'Achievements' }).click();
	await expect(page.getByLabel('Achievement points')).toHaveValue('75');
	await expect(page.getByLabel('Hide from players until the game ends')).toBeChecked();

	await page.setViewportSize({ width: 412, height: 915 });
	await expect(page.getByRole('button', { name: 'Achievements' })).toBeVisible();
});

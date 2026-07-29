import { expect, test } from '@playwright/test';

test('owner can create a ruleset', async ({ page }) => {
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Set up the host' })).toBeVisible();

	await page.getByLabel('Username').fill('partyhost');
	await page.getByLabel('Display name').fill('Party Host');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();

	await expect(page.getByRole('navigation', { name: 'Management' })).toBeVisible();
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await expect(page.getByRole('heading', { name: 'Rulesets' })).toBeVisible();
	await page.getByRole('link', { name: 'New ruleset' }).first().click();

	await expect(page.getByRole('heading', { name: 'Overview and limits' })).toBeVisible();
	await page.getByLabel('Stable slug').fill('party-test');
	await page.getByLabel('Name').fill('Party Test');
	await page.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/metadata$/);

	await page.reload();
	await expect(page.getByLabel('Stable slug')).toHaveValue('party-test');
	await expect(page.getByLabel('Name')).toHaveValue('Party Test');
});

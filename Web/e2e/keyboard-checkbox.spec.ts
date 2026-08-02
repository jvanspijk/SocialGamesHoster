import { expect, test } from '@playwright/test';

test('display setting can be changed with the keyboard', async ({ page }) => {
	await page.goto('/');
	await page.getByLabel('Username').fill('keyboardowner');
	await page.getByLabel('Display name').fill('Keyboard Owner');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();

	await page.getByRole('link', { name: 'Settings', exact: true }).click();
	await page.getByRole('link', { name: 'Display', exact: true }).click();
	const largeText = page.getByRole('checkbox', { name: /^Large text/ });
	await largeText.focus();
	await page.keyboard.press('Space');
	await expect(largeText).toBeChecked();
});

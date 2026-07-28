import { expect, test } from '@playwright/test';

async function expectNoHorizontalOverflow(page: import('@playwright/test').Page) {
	const overflow = await page.evaluate(() => ({
		page: Math.max(0, document.documentElement.scrollWidth - window.innerWidth),
		controls: Array.from(document.querySelectorAll('button, a, input, select, textarea'))
			.filter((element) => {
				const rect = element.getBoundingClientRect();
				return (
					rect.width > 0 &&
					rect.height > 0 &&
					(rect.left < -1 || rect.right > window.innerWidth + 1)
				);
			})
			.map((element) => element.getAttribute('aria-label') || element.textContent?.trim())
	}));
	expect(overflow).toEqual({ page: 0, controls: [] });
}

test('owner setup opens the replacement ruleset workspace', async ({ page }) => {
	await page.setViewportSize({ width: 320, height: 568 });
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Set up the host' })).toBeVisible();
	await expectNoHorizontalOverflow(page);

	await page.getByLabel('Username').fill('partyhost');
	await page.getByLabel('Display name').fill('Party Host');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();

	await expect(page.getByRole('navigation', { name: 'Management' })).toBeVisible();
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await expect(page.getByRole('heading', { name: 'Rulesets' })).toBeVisible();
	await page.getByRole('link', { name: 'New ruleset' }).first().click();

	await expect(page.getByRole('heading', { name: 'Overview and limits' })).toBeVisible();
	await page.getByLabel('Stable slug').fill('party-test');
	await page.getByLabel('Name').fill('Party Test');
	await page.getByRole('button', { name: 'Rewards', exact: true }).click();
	await page.getByRole('button', { name: 'Add achievement' }).click();
	await page.getByLabel('Achievement points').fill('75');
	await page.getByLabel('Hide from players until the game ends').check();

	await expect(page.getByLabel('Achievement points')).toHaveValue('75');
	await expect(page.getByRole('button', { name: 'Advanced JSON', exact: true })).toBeVisible();
	await expect(page.getByLabel('Complete ruleset JSON')).not.toBeVisible();

	await page.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/metadata$/);
	await expect(page.getByText(/Ruleset saved (with validation issues|and ready)/)).toBeVisible();

	await page.reload();
	await page.getByRole('button', { name: 'Rewards', exact: true }).click();
	await expect(page.getByLabel('Achievement points')).toHaveValue('75');
	await expect(page.getByLabel('Hide from players until the game ends')).toBeChecked();

	for (const viewport of [
		{ width: 320, height: 568 },
		{ width: 390, height: 844 },
		{ width: 1440, height: 900 }
	]) {
		await page.setViewportSize(viewport);
		await expectNoHorizontalOverflow(page);
	}
	await page.setViewportSize({ width: 412, height: 915 });
	await expect(page.getByRole('button', { name: /sections · rewards/i })).toBeVisible();
});

import { expect, test } from '@playwright/test';

test('ruleset workflow supports the accessibility display and keyboard matrix', async ({
	page
}) => {
	test.setTimeout(60_000);
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await page.setViewportSize({ width: 320, height: 720 });
	await page.goto('/');

	if (await page.getByRole('heading', { name: 'Set up the host' }).isVisible()) {
		await page.getByLabel('Username').fill('keyboardowner');
		await page.getByLabel('Display name').fill('Keyboard Owner');
		await page.getByLabel(/^Password/).fill('correct-horse-battery');
		await page.getByLabel('I understand and trust this local network.').check();
		await page.getByRole('button', { name: 'Create owner' }).click();
		await expect(page.getByRole('navigation', { name: 'Management' })).toBeVisible();
	} else {
		await page.getByRole('link', { name: 'Click here if you are a game master' }).click();
		await page.getByRole('textbox', { name: /^Username$/ }).fill('partyhost');
		await page.getByLabel(/^Password/).fill('correct-horse-battery');
		await page.getByRole('button', { name: 'Sign in' }).click();
		await expect(page.getByRole('heading', { name: 'Sign in' })).not.toBeVisible();
	}

	await page.goto('/admin/settings/display');
	const largeText = page.getByRole('checkbox', { name: /^Large text/ });
	await largeText.focus();
	await page.keyboard.press('Space');
	await expect(largeText).toBeChecked();
	const highContrast = page.getByRole('checkbox', { name: /^High contrast/ });
	await highContrast.focus();
	await page.keyboard.press('Space');
	await expect(highContrast).toBeChecked();
	await expect(page.locator('html')).toHaveAttribute('data-text-size', 'large');
	await expect(page.locator('html')).toHaveAttribute('data-contrast', 'high');
	await expect
		.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1))
		.toBe(true);

	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await page.getByRole('link', { name: /Echo Location/ }).click();
	await expect(page.getByRole('main')).toBeVisible();
	await expect(page.getByRole('link', { name: 'Rulesets', exact: true })).toBeVisible();
	await expect(page.getByText(/Saved · (Valid|Invalid)/)).toHaveAttribute('aria-live', 'polite');
	const preview = page.getByRole('button', { name: 'Preview', exact: true });
	await preview.focus();
	await page.keyboard.press('Enter');
	const dialog = page.getByRole('dialog', { name: 'Preview ruleset' });
	await expect(dialog).toBeVisible();
	await expect(dialog.getByRole('group', { name: 'Preview type' })).toBeVisible();
	await expect(dialog.getByRole('button', { name: 'Close Preview ruleset' })).toBeFocused();
	await expect
		.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1))
		.toBe(true);
	await page.keyboard.press('Escape');
	await expect(preview).toBeFocused();

	// A 320 CSS-pixel layout is the reflow viewport produced by 200% browser
	// zoom on a 640-pixel window; CSS `zoom` would scale the document without
	// shrinking its layout viewport and is not equivalent.
	await page.setViewportSize({ width: 320, height: 720 });
	await expect
		.poll(() => page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches))
		.toBe(true);
});

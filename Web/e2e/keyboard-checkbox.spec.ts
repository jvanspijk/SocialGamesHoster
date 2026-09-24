import { expect, test } from '@playwright/test';

test('ruleset workflow supports the accessibility display and keyboard matrix', async ({
	page
}) => {
	test.setTimeout(60_000);
	await page.emulateMedia({ reducedMotion: 'reduce' });
	await page.setViewportSize({ width: 320, height: 720 });
	await page.goto('/');
	await expect(
		page
			.getByRole('heading', { name: 'Set up the app' })
			.or(page.getByRole('link', { name: 'Game Master sign in' }))
	).toBeVisible();

	if (await page.getByRole('heading', { name: 'Set up the app' }).isVisible()) {
		await expect
			.poll(() =>
				page.getByLabel('Username').evaluate((input) => input.getBoundingClientRect().height)
			)
			.toBeGreaterThanOrEqual(44);
		await page.getByLabel('Username').fill('keyboardowner');
		await page.getByLabel('Display name').fill('Keyboard Owner');
		await page.getByLabel(/^Password/).fill('correct-horse-battery');
		await page.getByLabel('I understand and trust this local network.').check();
		await page.getByRole('button', { name: 'Create owner' }).click();
		await expect(page.getByRole('navigation', { name: 'Management' })).toBeVisible();
	} else {
		await page.getByRole('link', { name: 'Game Master sign in' }).click();
		await page.getByRole('textbox', { name: /^Username$/ }).fill('partyhost');
		await page.getByLabel(/^Password/).fill('correct-horse-battery');
		await page.getByRole('button', { name: 'Sign in' }).click();
		await expect(page.getByRole('heading', { name: 'Sign in' })).not.toBeVisible();
	}

	await page.goto('/admin/approvals');
	const search = page.getByRole('searchbox', { name: 'Search profiles' });
	await search.focus();
	await expect(search.locator('..')).toHaveCSS('outline-width', '3px');

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
		.poll(() =>
			page.locator('html').evaluate((html) => parseFloat(getComputedStyle(html).fontSize))
		)
		.toBeGreaterThan(18);
	await expect(page.locator('.immersive-shell')).toHaveCSS('background-image', 'none');
	await expect
		.poll(() => page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1))
		.toBe(true);

	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await page.getByRole('link', { name: /Echo Location/ }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/metadata$/);
	await expect(page.getByRole('link', { name: 'Rulesets', exact: true })).toBeVisible();
	await expect(page.getByText('All changes saved', { exact: true })).toHaveAttribute(
		'aria-live',
		'polite'
	);
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

	await page.getByRole('textbox', { name: /^Name/ }).fill('Keyboard editor changes');
	await expect(page.getByText('Unsaved changes', { exact: true })).toHaveAttribute(
		'aria-live',
		'polite'
	);
	await expect
		.poll(() =>
			page.evaluate(() =>
				Object.keys(localStorage).some((key) =>
					key.startsWith('social-games-hoster:ruleset-working-copy:')
				)
			)
		)
		.toBe(true);
	await page.reload();
	await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
	await expect(page.getByRole('textbox', { name: /^Name/ })).toHaveValue('Keyboard editor changes');
	await page.getByRole('button', { name: /^Overview/ }).click();
	await expect(
		page
			.getByRole('dialog', { name: 'Ruleset overview' })
			.getByRole('heading', { name: 'Overview', exact: true })
	).toBeVisible();
	await expect(page.getByText('Saved version: Game readiness')).not.toBeVisible();
	await page.keyboard.press('Escape');
	for (const width of [1280, 1600, 320]) {
		await page.setViewportSize({ width, height: 900 });
		await expect
			.poll(() =>
				page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)
			)
			.toBe(true);
		if (width === 1600)
			await page.screenshot({ path: 'test-results/ruleset-editor-desktop.png', fullPage: true });
	}

	// A 320 CSS-pixel layout is the reflow viewport produced by 200% browser
	// zoom on a 640-pixel window; CSS `zoom` would scale the document without
	// shrinking its layout viewport and is not equivalent.
	await page.setViewportSize({ width: 320, height: 720 });
	await expect
		.poll(() => page.evaluate(() => matchMedia('(prefers-reduced-motion: reduce)').matches))
		.toBe(true);
});

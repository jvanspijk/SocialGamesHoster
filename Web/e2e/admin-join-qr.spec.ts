import { expect, test } from '@playwright/test';

test('admin header shows the player join QR code on desktop and mobile', async ({ page }) => {
	await page.goto('/');
	await page.getByLabel('Username').fill('partyhost');
	await page.getByLabel('Display name').fill('Party Host');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();

	for (const width of [1280, 375]) {
		await page.setViewportSize({ width, height: 800 });
		const trigger = page.getByRole('button', { name: 'Show player join QR code' });
		await expect(trigger).toBeVisible();
		await expect(page.getByText(/^connected$/i)).toHaveCount(0);
		await trigger.click();

		const dialog = page.getByRole('dialog', { name: 'Player join QR code' });
		await expect(dialog).toBeVisible();
		const image = dialog.getByRole('img', { name: 'QR code for the player join page' });
		await expect(image).toBeVisible();
		await expect
			.poll(() => image.evaluate((element: HTMLImageElement) => element.naturalWidth))
			.toBeGreaterThan(0);
		const box = await dialog.boundingBox();
		expect(box).not.toBeNull();
		expect(box!.x).toBeGreaterThanOrEqual(0);
		expect(box!.x + box!.width).toBeLessThanOrEqual(width);
		await page.keyboard.press('Escape');
		await expect(dialog).toBeHidden();
		await expect(trigger).toBeFocused();
	}
});

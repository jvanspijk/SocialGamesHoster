import { expect, test } from '@playwright/test';

test('owner can create a ruleset', async ({ page }) => {
	test.setTimeout(60_000);
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
	await page.getByRole('link', { name: 'Create ruleset' }).first().click();

	await expect(page.getByRole('heading', { name: 'Create ruleset' })).toBeVisible();
	await page.getByLabel('Ruleset name').fill('Party Test');
	await page.getByRole('button', { name: 'Create ruleset' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/metadata$/);

	await page.getByRole('textbox', { name: /^Name$/ }).fill('Recovered Party Test');
	await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
	await page.waitForTimeout(600);
	await page.reload();
	await expect(page.getByRole('textbox', { name: /^Name$/ })).toHaveValue('Recovered Party Test');
	await expect(page.getByText('Recovered changes', { exact: true })).toBeVisible();

	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	const leaveDialog = page.getByRole('dialog', { name: 'Leave with unsaved changes?' });
	await expect(leaveDialog).toBeVisible();
	await leaveDialog.getByRole('button', { name: 'Keep editing' }).click();
	await expect(page.getByRole('textbox', { name: /^Name$/ })).toHaveValue('Recovered Party Test');
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await leaveDialog.getByRole('button', { name: 'Save and leave' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await page.getByLabel('Description').fill('Discard this change');
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await leaveDialog.getByRole('button', { name: 'Discard and leave' }).click();
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByLabel('Description')).toHaveValue('');

	const actions = page.getByRole('button', { name: 'Actions', exact: true });
	await actions.click();
	const actionsDialog = page.getByRole('dialog', { name: 'Ruleset actions' });
	await expect(actionsDialog).toBeVisible();
	await expect(page.getByRole('button', { name: 'Close Ruleset actions' })).toBeFocused();
	await page.keyboard.press('Escape');
	await expect(actionsDialog).not.toBeVisible();
	await expect(actions).toBeFocused();

	await page.setViewportSize({ width: 390, height: 844 });
	const sections = page.getByRole('button', { name: /Sections/ });
	await sections.click();
	const sheet = page.getByRole('dialog', { name: 'Ruleset sections' });
	await expect(sheet).toBeVisible();
	await expect(page.getByRole('button', { name: 'Close Ruleset sections' })).toBeFocused();
	const phoneSheet = await sheet.boundingBox();
	expect(phoneSheet).not.toBeNull();
	expect(phoneSheet?.x).toBe(0);
	expect(phoneSheet?.y).toBe(0);
	expect(phoneSheet?.width).toBe(390);
	expect(phoneSheet?.height).toBe(844);

	await page.setViewportSize({ width: 1280, height: 800 });
	const desktopSheet = await sheet.boundingBox();
	expect(desktopSheet).not.toBeNull();
	expect(desktopSheet?.width).toBeLessThan(600);
	expect(desktopSheet?.height).toBeLessThan(800);
	expect(desktopSheet?.x).toBeGreaterThan(600);
	expect(desktopSheet?.y).toBeGreaterThan(0);

	await page.setViewportSize({ width: 390, height: 844 });
	await page.keyboard.press('Escape');
	await expect(sheet).not.toBeVisible();
	await expect(sections).toBeFocused();

	await sections.click();
	await sheet.getByRole('button', { name: /^Player setup/ }).click();
	await page.getByRole('button', { name: 'Add band', exact: true }).click();

	const roleSlots = page
		.locator('.content-header')
		.filter({ has: page.getByText('Role slots', { exact: true }) });
	await expect(roleSlots).toBeVisible();

	await page.setViewportSize({ width: 700, height: 844 });
	await expect(roleSlots).toHaveCSS('flex-direction', 'column');
	const roleSlotsTitle = roleSlots.getByText('Role slots', { exact: true });
	const addSlot = roleSlots.getByRole('button', { name: 'Add slot', exact: true });
	const titleBox = await roleSlotsTitle.boundingBox();
	const actionBox = await addSlot.boundingBox();
	expect(titleBox).not.toBeNull();
	expect(actionBox).not.toBeNull();
	expect(actionBox!.y).toBeGreaterThan(titleBox!.y);
});

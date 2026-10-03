import { expect, test } from './fixtures';

test('owner completes the ruleset lifecycle with recovery, assets, previews, and a new game', async ({
	page
}) => {
	test.setTimeout(120_000);
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Set up the app' })).toBeVisible();

	await page.getByLabel('Username').fill('partyhost');
	await page.getByLabel('Display name').fill('Party Host');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();

	await expect(page.getByRole('navigation', { name: 'Management' })).toBeVisible();
	await page
		.getByRole('navigation', { name: 'Management' })
		.getByRole('link', { name: 'Rulesets', exact: true })
		.click();
	await expect(page.getByRole('heading', { name: 'Rulesets' })).toBeVisible();
	await page.getByRole('link', { name: 'Create ruleset' }).first().click();

	await expect(page.getByRole('heading', { name: 'Create ruleset' })).toBeVisible();
	await page.getByLabel('Ruleset name').fill('Party Test');
	await page.getByRole('button', { name: 'Create ruleset' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/metadata$/);

	await page.getByRole('textbox', { name: /^Name$/ }).fill('Recovered Party Test');
	await page.getByLabel('Maximum players').fill('3');
	await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
	await page.waitForTimeout(600);
	await page.reload();
	await expect(page.getByRole('textbox', { name: /^Name$/ })).toHaveValue('Recovered Party Test');
	await page.getByRole('button', { name: 'Upload new' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/assets$/);
	const coverChooser = page.waitForEvent('filechooser');
	await page.getByRole('button', { name: 'Upload image' }).click();
	await (
		await coverChooser
	).setFiles({
		name: 'party-cover.png',
		mimeType: 'image/png',
		buffer: Buffer.from(
			'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL0WQAAAABJRU5ErkJggg==',
			'base64'
		)
	});
	await expect(page.getByRole('textbox', { name: 'Display name' })).toHaveValue('party-cover.png');
	await page.getByRole('textbox', { name: 'Display name' }).fill('Party cover');
	await page
		.getByRole('textbox', { name: 'Description', exact: true })
		.fill('Friends gathered for a game');
	await page
		.getByRole('complementary', { name: 'Item details' })
		.getByRole('button', { name: 'Save', exact: true })
		.click();
	await page
		.getByRole('navigation', { name: 'Ruleset sections' })
		.getByRole('button', { name: /^Basics/ })
		.click();
	const rulesetCover = page.getByRole('region', { name: 'Ruleset cover' }).getByRole('combobox');
	await rulesetCover.selectOption({ label: 'Party cover' });
	await expect(rulesetCover.locator('option:checked')).toHaveText('Party cover');

	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	const leaveDialog = page.getByRole('dialog', { name: 'Leave with unsaved changes?' });
	await expect(leaveDialog).toBeVisible();
	await leaveDialog.getByRole('button', { name: 'Keep editing' }).click();
	await expect(page.getByRole('textbox', { name: /^Name$/ })).toHaveValue('Recovered Party Test');
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await leaveDialog.getByRole('button', { name: 'Save and leave' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(rulesetCover.locator('option:checked')).toHaveText('Party cover');
	await page.getByLabel(/^Description/).fill('Discard this change');
	const discardPattern = '**/api/app/v1/rulesets/*/edit-session/*';
	await page.route(discardPattern, async (route) => {
		if (route.request().method() === 'DELETE') await route.abort();
		else await route.continue();
	});
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await leaveDialog.getByRole('button', { name: 'Discard and leave' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await expect(
		page.getByText(
			/Your unsaved changes were discarded\. Some uploaded assets could not be cleaned up and will expire automatically\./
		)
	).toBeVisible();
	await page.unroute(discardPattern);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByLabel(/^Description/)).toHaveValue('');

	const previewAction = page.getByRole('button', { name: 'Preview', exact: true });
	await previewAction.click();
	const previewSheet = page.getByRole('dialog', { name: 'Preview ruleset' });
	await expect(previewSheet).toBeVisible();
	await expect(previewSheet.getByText('Previewing the saved ruleset')).toBeVisible();
	await page.keyboard.press('Escape');
	await expect(previewSheet).not.toBeVisible();
	await expect(previewAction).toBeFocused();

	const deleteAction = page.getByRole('button', { name: 'Delete ruleset', exact: true });
	await deleteAction.click();
	const deleteDialog = page.getByRole('dialog', { name: 'Delete ruleset?' });
	await expect(deleteDialog).toBeVisible();
	await expect(deleteDialog.getByRole('button', { name: 'Close Delete ruleset?' })).toBeFocused();
	await page.keyboard.press('Escape');
	await expect(deleteDialog).not.toBeVisible();
	await expect(deleteAction).toBeFocused();

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
	expect(desktopSheet?.width).toBeLessThanOrEqual(640);
	expect(desktopSheet?.height).toBeLessThan(800);
	expect(desktopSheet?.x).toBeGreaterThan(600);
	expect(desktopSheet?.y).toBeGreaterThan(0);

	await page.setViewportSize({ width: 390, height: 844 });
	await page.keyboard.press('Escape');
	await expect(sheet).not.toBeVisible();
	await expect(sections).toBeFocused();

	await page.setViewportSize({ width: 1280, height: 800 });
	const sectionRail = page.getByRole('navigation', { name: 'Ruleset sections' });
	await sectionRail.getByRole('button', { name: /^Teams/ }).click();
	const teamsEditor = page.getByRole('region', { name: 'Teams' });
	await teamsEditor.getByRole('button', { name: 'Add teams' }).click();
	await teamsEditor.getByRole('textbox', { name: 'Name', exact: true }).fill('Village');

	await sectionRail.getByRole('button', { name: /^Roles and abilities/ }).click();
	const rolesEditor = page.getByRole('region', { name: 'Roles' });
	await rolesEditor.getByRole('button', { name: 'Add roles' }).click();
	await rolesEditor.getByRole('textbox', { name: 'Name', exact: true }).fill('Villager');
	await rolesEditor
		.getByRole('textbox', { name: 'Description', exact: true })
		.fill('Keep the village safe.');
	await rolesEditor.getByLabel('Win condition').fill('Find every threat.');

	await sectionRail.getByRole('button', { name: /^Assets/ }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets\/[^/]+\/edit\/assets$/);
	await page.getByRole('button', { name: /Party cover/ }).click();
	const replacementChooser = page.waitForEvent('filechooser');
	await page.getByRole('button', { name: 'Replace', exact: true }).click();
	await (
		await replacementChooser
	).setFiles({
		name: 'replacement-cover.png',
		mimeType: 'image/png',
		buffer: Buffer.from(
			'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL0WQAAAABJRU5ErkJggg==',
			'base64'
		)
	});
	await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();

	await page.getByRole('button', { name: 'Save', exact: true }).first().click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();

	await page.getByRole('button', { name: 'Preview', exact: true }).click();
	const completedPreview = page.getByRole('dialog', { name: 'Preview ruleset' });
	await expect(completedPreview.getByRole('heading', { name: 'Villager' })).toBeVisible();
	await completedPreview.getByRole('button', { name: 'Assets', exact: true }).click();
	await expect(completedPreview.getByRole('heading', { name: 'In the game' })).toBeVisible();
	await expect(
		completedPreview.getByRole('heading', { name: 'Recovered Party Test' })
	).toBeVisible();
	await page.keyboard.press('Escape');
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await page.getByRole('link', { name: 'Games', exact: true }).click();
	await page.getByRole('button', { name: 'New game' }).first().click();
	const gameDialog = page.getByRole('dialog', { name: 'New game' });
	await gameDialog.getByLabel('Game name').fill('Recovered Ruleset Game');
	await gameDialog.getByLabel('Ruleset').selectOption({ label: 'Recovered Party Test' });
	await gameDialog.getByRole('button', { name: 'Create game' }).click();
	await expect(page).toHaveURL(/\/admin\/games\/[^/]+\/overview$/);
	await expect(page.getByRole('heading', { name: 'QR Code' })).toBeVisible();
	await page.goto('/admin/games');
	await expect(page.getByRole('columnheader', { name: 'Game' })).toBeVisible();
	await expect(page.getByRole('cell', { name: '0/3' })).toBeVisible();
	await page.setViewportSize({ width: 390, height: 844 });
	await expect(page.getByRole('link', { name: 'Recovered Ruleset Game' })).toBeVisible();
	expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth)).toBe(
		true
	);
	await page.setViewportSize({ width: 1280, height: 800 });
	await page.getByRole('link', { name: 'Recovered Ruleset Game' }).click();
	await page.getByRole('button', { name: 'Cancel game' }).click();
	const cancelDialog = page.getByRole('dialog', { name: 'Cancel game?' });
	await cancelDialog.getByRole('button', { name: 'Cancel game', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/games$/);
	await expect(page.getByRole('link', { name: 'Recovered Ruleset Game' })).not.toBeVisible();

	await page.goto('/admin/rulesets');
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByText('All changes saved', { exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Delete ruleset', exact: true }).click();
	const finalDeleteDialog = page.getByRole('dialog', { name: 'Delete ruleset?' });
	await finalDeleteDialog.getByRole('button', { name: 'Delete ruleset', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await expect(page.getByRole('link', { name: /Recovered Party Test/ })).not.toBeVisible();
});

test('announcement composer sends ruleset and one-off media to a recipient', async ({
	browser,
	page
}) => {
	test.setTimeout(180_000);
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Set up the app' })).toBeVisible();
	await page.getByLabel('Username').fill('partyhost');
	await page.getByLabel('Display name').fill('Party Host');
	await page.getByLabel('Password').fill('correct-horse-battery');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();

	await page
		.getByRole('navigation', { name: 'Management' })
		.getByRole('link', { name: 'Games', exact: true })
		.click();
	await page.getByRole('button', { name: 'New game' }).first().click();
	const gameDialog = page.getByRole('dialog', { name: 'New game' });
	await gameDialog.getByLabel('Game name').fill('Announcement Browser Test');
	await gameDialog.getByLabel('Ruleset').selectOption({ label: 'Echo Location' });
	await gameDialog.getByRole('button', { name: 'Create game' }).click();
	await expect(page).toHaveURL(/\/admin\/games\/[^/]+\/overview$/);
	const gameUrl = page.url();
	await expect(page.getByRole('heading', { name: 'QR Code' })).toBeVisible();

	const playerContext = await browser.newContext();
	const player = await playerContext.newPage();
	const nonRecipientContext = await browser.newContext();
	const nonRecipient = await nonRecipientContext.newPage();
	const roleAssignmentContext = await browser.newContext();
	const roleAssignmentPlayer = await roleAssignmentContext.newPage();
	try {
		await player.goto('/');
		await player.getByLabel('Profile name').fill('Browser Player');
		await player.getByRole('button', { name: 'Join game' }).click();
		await expect(player.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();

		await page.getByRole('button', { name: /Entry requests, 1 waiting/ }).click();
		const request = page.getByRole('article').filter({ hasText: 'Browser Player' });
		await expect(request).toBeVisible();
		await request.getByRole('button', { name: 'Approve' }).click();
		await expect(player.getByRole('link', { name: 'Current game' })).toBeVisible();
		await player.goto('/play/party');
		await expect(player.getByText('Player 1', { exact: true })).toBeVisible();
		await page.goto(gameUrl);

		await page.getByRole('button', { name: 'New announcement' }).first().click();
		let announcement = page.getByRole('dialog', { name: 'New announcement' });
		await announcement.getByLabel('Announcement message').fill('Existing media announcement');
		await announcement
			.getByRole('group', { name: 'Image (optional)' })
			.getByLabel('Choose from ruleset')
			.check();
		await announcement
			.getByLabel('Ruleset image')
			.selectOption({ label: 'echo-location-cover.webp' });
		await announcement
			.getByRole('group', { name: 'Audio (optional)' })
			.getByLabel('Choose from ruleset')
			.check();
		await announcement.getByLabel('Ruleset audio').selectOption({ label: 'course-clear.ogg' });
		await announcement.getByRole('button', { name: 'Send announcement' }).click();
		await expect(page.getByText('Announcement sent.')).toBeVisible();
		await expect(player.getByText('Existing media announcement')).toBeVisible();
		await expect(
			player.getByAltText(
				'A stylized submarine travels through dark blue water while sonar rings reveal rocks and distant hazards.'
			)
		).toBeVisible();
		await expect(player.locator('audio')).toBeAttached();
		await player.getByRole('button', { name: 'Acknowledge' }).click();
		await expect(player.getByText('Existing media announcement')).not.toBeVisible();

		let nonRecipientAuthorization = '';
		nonRecipient.on('request', (request) => {
			if (request.url().includes('/api/app/v1/')) {
				nonRecipientAuthorization = request.headers().authorization ?? nonRecipientAuthorization;
			}
		});
		await nonRecipient.goto('/');
		await nonRecipient.getByLabel('Profile name').fill('Other Browser Player');
		await nonRecipient.getByRole('button', { name: 'Join game' }).click();
		await expect(nonRecipient.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();
		await page.getByRole('button', { name: /Entry requests, 1 waiting/ }).click();
		const otherRequest = page.getByRole('article').filter({ hasText: 'Other Browser Player' });
		await expect(otherRequest).toBeVisible();
		await otherRequest.getByRole('button', { name: 'Approve' }).click();
		await expect(nonRecipient).toHaveURL(/\/play(?:\/party)?$/);
		await expect.poll(() => nonRecipientAuthorization).not.toBe('');
		await page.goto(gameUrl);

		await roleAssignmentPlayer.goto('/');
		await roleAssignmentPlayer.getByLabel('Profile name').fill('Role Assignment Player');
		await roleAssignmentPlayer.getByRole('button', { name: 'Join game' }).click();
		await expect(
			roleAssignmentPlayer.getByRole('heading', { name: 'Awaiting approval' })
		).toBeVisible();
		await page.getByRole('button', { name: /Entry requests, 1 waiting/ }).click();
		const roleAssignmentRequest = page
			.getByRole('article')
			.filter({ hasText: 'Role Assignment Player' });
		await expect(roleAssignmentRequest).toBeVisible();
		await roleAssignmentRequest.getByRole('button', { name: 'Approve' }).click();
		await expect(roleAssignmentPlayer).toHaveURL(/\/play(?:\/party)?$/);
		await page.goto(gameUrl);

		await page.getByRole('link', { name: 'Players', exact: true }).click();
		await page.getByLabel('Role for Browser Player').selectOption({ label: 'Lookout' });
		await page
			.getByLabel('Role for Other Browser Player')
			.selectOption({ label: 'Sonar Operator' });
		const saveRoles = page.getByRole('button', { name: 'Save roles' });
		await expect(saveRoles).toBeEnabled();
		await saveRoles.click();
		await expect(page.getByText('Role assignments saved.')).toBeVisible();
		await expect(page.getByLabel('Role for Browser Player')).toHaveValue('lookout');
		await expect(page.getByLabel('Role for Other Browser Player')).toHaveValue('sonar_operator');
		await expect(page.getByLabel('Role for Role Assignment Player')).toHaveValue('');
		await page.getByRole('link', { name: 'Overview', exact: true }).click();
		await page.getByRole('link', { name: 'Players', exact: true }).click();
		await expect(page.getByLabel('Role for Browser Player')).toHaveValue('lookout');
		await expect(page.getByLabel('Role for Other Browser Player')).toHaveValue('sonar_operator');
		await expect(page.getByLabel('Role for Role Assignment Player')).toHaveValue('');
		await page.reload();
		await expect(page.getByLabel('Role for Browser Player')).toHaveValue('lookout');
		await expect(page.getByLabel('Role for Other Browser Player')).toHaveValue('sonar_operator');
		await expect(page.getByLabel('Role for Role Assignment Player')).toHaveValue('');

		await page.goto(gameUrl);

		await page.getByRole('button', { name: 'New announcement' }).first().click();
		announcement = page.getByRole('dialog', { name: 'New announcement' });
		await announcement.getByLabel('Announcement message').fill('Uploaded media announcement');
		await announcement.getByLabel('Recipients').selectOption('player');
		const browserPlayerId = await announcement
			.getByLabel('Player')
			.locator('option')
			.filter({ hasText: /^Player \d+ · Browser Player$/ })
			.getAttribute('value');
		expect(browserPlayerId).not.toBeNull();
		await announcement
			.getByRole('combobox', { name: 'Player', exact: true })
			.selectOption(browserPlayerId!);
		await announcement
			.getByRole('group', { name: 'Image (optional)' })
			.getByLabel('Upload for this announcement')
			.check();
		await announcement.getByLabel('Image file').setInputFiles({
			name: 'one-off.png',
			mimeType: 'image/png',
			buffer: Buffer.from(
				'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL0WQAAAABJRU5ErkJggg==',
				'base64'
			)
		});
		let failedOnce = false;
		const announcementPattern = '**/api/app/v1/games/*/announcements';
		await page.route(announcementPattern, async (route) => {
			if (!failedOnce && route.request().method() === 'POST') {
				failedOnce = true;
				await route.abort('failed');
				return;
			}
			await route.continue();
		});
		await announcement.getByRole('button', { name: 'Send announcement' }).click();
		await expect(page.getByText('The app returned an unexpected response.')).toBeVisible();
		await expect(announcement).toBeVisible();
		await expect(announcement.getByLabel('Announcement message')).toHaveValue(
			'Uploaded media announcement'
		);
		await page.unroute(announcementPattern);
		const mediaRequests: Array<{ url: string; authorization: string }> = [];
		player.on('request', (request) => {
			if (/\/announcements\/[^/]+\/media\/(image|audio)$/.test(new URL(request.url()).pathname)) {
				mediaRequests.push({
					url: request.url(),
					authorization: request.headers().authorization ?? ''
				});
			}
		});
		await announcement.getByRole('button', { name: 'Send announcement' }).click();
		await expect(page.getByText('Announcement sent.')).toBeVisible();
		await expect(player.getByText('Uploaded media announcement')).toBeVisible();
		await expect(player.locator('img').last()).toBeAttached();
		await expect(nonRecipient.getByText('Uploaded media announcement')).not.toBeVisible();
		await expect.poll(() => mediaRequests.length).toBeGreaterThanOrEqual(1);

		const deniedMedia = await nonRecipientContext.request.get(mediaRequests[0].url, {
			headers: { Authorization: nonRecipientAuthorization }
		});
		expect(deniedMedia.status()).toBe(403);
	} finally {
		await playerContext.close();
		await nonRecipientContext.close();
		await roleAssignmentContext.close();
	}
});

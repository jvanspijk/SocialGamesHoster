import { expect, test } from '@playwright/test';

function oneSecondWav() {
	const sampleRate = 8_000;
	const samples = Buffer.alloc(sampleRate, 128);
	const wav = Buffer.alloc(44 + samples.length);
	wav.write('RIFF', 0);
	wav.writeUInt32LE(wav.length - 8, 4);
	wav.write('WAVE', 8);
	wav.write('fmt ', 12);
	wav.writeUInt32LE(16, 16);
	wav.writeUInt16LE(1, 20);
	wav.writeUInt16LE(1, 22);
	wav.writeUInt32LE(sampleRate, 24);
	wav.writeUInt32LE(sampleRate, 28);
	wav.writeUInt16LE(1, 32);
	wav.writeUInt16LE(8, 34);
	wav.write('data', 36);
	wav.writeUInt32LE(samples.length, 40);
	samples.copy(wav, 44);
	return wav;
}

test('owner completes the ruleset lifecycle with recovery, media, previews, and a new game', async ({
	page
}) => {
	test.setTimeout(120_000);
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
	await page.getByLabel('Maximum players').fill('3');
	await expect(page.getByText('Unsaved changes', { exact: true })).toBeVisible();
	await page.waitForTimeout(600);
	await page.reload();
	await expect(page.getByRole('textbox', { name: /^Name$/ })).toHaveValue('Recovered Party Test');
	await expect(page.getByText('Recovered changes', { exact: true })).toBeVisible();
	await page.getByRole('textbox', { name: /Media name/ }).fill('Party cover');
	await page
		.getByRole('textbox', { name: /Image description/ })
		.fill('Friends gathered for a game');
	await page.locator('input[type="file"]').setInputFiles({
		name: 'party-cover.png',
		mimeType: 'image/png',
		buffer: Buffer.from(
			'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVQIHWP4z8DwHwAFgAI/ScL0WQAAAABJRU5ErkJggg==',
			'base64'
		)
	});
	await expect(page.getByLabel('Ruleset cover').locator('option:checked')).toHaveText(
		'Party cover'
	);

	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	const leaveDialog = page.getByRole('dialog', { name: 'Leave with unsaved changes?' });
	await expect(leaveDialog).toBeVisible();
	await leaveDialog.getByRole('button', { name: 'Keep editing' }).click();
	await expect(page.getByRole('textbox', { name: /^Name$/ })).toHaveValue('Recovered Party Test');
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await leaveDialog.getByRole('button', { name: 'Save and leave' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByLabel('Ruleset cover').locator('option:checked')).toHaveText(
		'Party cover'
	);
	await page.getByLabel('Description', { exact: true }).fill('Discard this change');
	const discardPattern = '**/api/app/v1/rulesets/*/edit-session/*';
	await page.route(discardPattern, async (route) => {
		if (route.request().method() === 'DELETE') await route.abort();
		else await route.continue();
	});
	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await leaveDialog.getByRole('button', { name: 'Discard and leave' }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await expect(
		page.getByText(/The local working copy was discarded\. Host cleanup could not be confirmed/)
	).toBeVisible();
	await page.unroute(discardPattern);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByLabel('Description', { exact: true })).toHaveValue('');

	const previewAction = page.getByRole('button', { name: 'Preview', exact: true });
	await previewAction.click();
	const previewSheet = page.getByRole('dialog', { name: 'Preview ruleset' });
	await expect(previewSheet).toBeVisible();
	await expect(previewSheet.getByText('Previewing the saved ruleset')).toBeVisible();
	await previewSheet.getByRole('button', { name: 'Player setup', exact: true }).click();
	await expect(previewSheet.getByRole('spinbutton', { name: 'Player count' })).toBeVisible();
	await expect(previewSheet.getByRole('heading', { name: 'Setup is not feasible' })).toBeVisible();
	await page.keyboard.press('Escape');
	await expect(previewSheet).not.toBeVisible();
	await expect(previewAction).toBeFocused();

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
	await rolesEditor.getByLabel('Maximum copies').fill('3');

	await sectionRail.getByRole('button', { name: /^Player setup/ }).click();
	await page.getByRole('button', { name: 'Add slot', exact: true }).click();
	await page.getByLabel('Number of players').fill('3');

	await sectionRail.getByRole('button', { name: /^Media/ }).click();
	await page.getByRole('button', { name: /Party cover/ }).click();
	const replacementChooser = page.waitForEvent('filechooser');
	await page.getByRole('button', { name: 'Replace everywhere' }).click();
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

	await page.getByRole('button', { name: 'Save', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await page.getByRole('link', { name: /Recovered Party Test/ }).click();
	await expect(page.getByText('Saved · Valid', { exact: true })).toBeVisible();

	await page.getByRole('button', { name: 'Preview', exact: true }).click();
	const completedPreview = page.getByRole('dialog', { name: 'Preview ruleset' });
	await expect(completedPreview.getByRole('heading', { name: 'Villager' })).toBeVisible();
	await completedPreview.getByRole('button', { name: 'Player setup', exact: true }).click();
	await expect(completedPreview.getByRole('heading', { name: 'Setup is feasible' })).toBeVisible();
	await completedPreview.getByRole('button', { name: 'Media', exact: true }).click();
	await expect(completedPreview.getByRole('heading', { name: 'In the game' })).toBeVisible();
	await expect(
		completedPreview.getByRole('heading', { name: 'Recovered Party Test' })
	).toBeVisible();
	await page.keyboard.press('Escape');
	await expect(
		page.locator('aside.readiness').getByText('A valid setup is available for 3 players.')
	).toBeVisible();

	await page.getByRole('link', { name: 'Rulesets', exact: true }).click();
	await page.getByRole('link', { name: 'Games', exact: true }).click();
	await page.getByRole('button', { name: 'New game' }).first().click();
	const gameDialog = page.getByRole('dialog', { name: 'New game' });
	await gameDialog.getByLabel('Game name').fill('Recovered Ruleset Game');
	await gameDialog.getByLabel('Ruleset').selectOption({ label: 'Recovered Party Test' });
	await gameDialog.getByRole('button', { name: 'Create game' }).click();
	await expect(page).toHaveURL(/\/admin\/games\/[^/]+\/overview$/);
	await expect(page.getByText('Player invitation')).toBeVisible();
	await page.getByRole('link', { name: 'Back to Games' }).click();
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
	await expect(page.getByText('Saved · Valid', { exact: true })).toBeVisible();
	await page.getByRole('button', { name: 'Actions', exact: true }).click();
	await page.getByRole('button', { name: 'Delete ruleset' }).click();
	const deleteDialog = page.getByRole('dialog', { name: 'Delete ruleset?' });
	await deleteDialog.getByRole('button', { name: 'Delete ruleset', exact: true }).click();
	await expect(page).toHaveURL(/\/admin\/rulesets$/);
	await expect(page.getByRole('link', { name: /Recovered Party Test/ })).not.toBeVisible();
});

test('announcement composer sends ruleset and one-off media to a recipient', async ({
	browser,
	page
}) => {
	test.setTimeout(90_000);
	await page.goto('/');
	if (await page.getByRole('heading', { name: 'Set up the host' }).isVisible()) {
		await page.getByLabel('Username').fill('partyhost');
		await page.getByLabel('Display name').fill('Party Host');
		await page.getByLabel('Password').fill('correct-horse-battery');
		await page.getByLabel('I understand and trust this local network.').check();
		await page.getByRole('button', { name: 'Create owner' }).click();
	} else {
		await page.getByRole('link', { name: 'Click here if you are a game master' }).click();
		await page.getByRole('textbox', { name: /^Username$/ }).fill('partyhost');
		await page.getByRole('textbox', { name: /^Password$/ }).fill('correct-horse-battery');
		await page.getByRole('button', { name: 'Sign in' }).click();
	}

	await page.getByRole('link', { name: 'Games', exact: true }).click();
	await page.getByRole('button', { name: 'New game' }).first().click();
	const gameDialog = page.getByRole('dialog', { name: 'New game' });
	await gameDialog.getByLabel('Game name').fill('Announcement Browser Test');
	await gameDialog.getByLabel('Ruleset').selectOption({ label: 'Echo Location' });
	await gameDialog.getByRole('button', { name: 'Create game' }).click();
	await expect(page).toHaveURL(/\/admin\/games\/[^/]+\/overview$/);
	const gameUrl = page.url();
	await expect(page.getByText('Player invitation')).toBeVisible();

	const playerContext = await browser.newContext();
	const player = await playerContext.newPage();
	const nonRecipientContext = await browser.newContext();
	const nonRecipient = await nonRecipientContext.newPage();
	try {
		await player.goto('/');
		await player.getByLabel('Profile name').fill('Browser Player');
		await player.getByRole('button', { name: 'Request entry' }).click();
		await expect(player.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();

		await page.getByRole('button', { name: /Entry requests, 1 waiting/ }).click();
		const request = page.getByRole('article').filter({ hasText: 'Browser Player' });
		await expect(request).toBeVisible();
		await request.getByRole('button', { name: 'Approve' }).click();
		await expect(player).toHaveURL(/\/play(?:\/party)?$/);
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
		await announcement.getByLabel('Image description').fill('The Echo Location cover art');
		await announcement
			.getByRole('group', { name: 'Audio (optional)' })
			.getByLabel('Choose from ruleset')
			.check();
		await announcement.getByLabel('Ruleset audio').selectOption({ label: 'course-clear.ogg' });
		await announcement.getByLabel('Audio alternative').fill('A course-clear signal');
		await announcement.getByRole('button', { name: 'Send announcement' }).click();
		await expect(page.getByText('Announcement sent.')).toBeVisible();
		await expect(player.getByText('Existing media announcement')).toBeVisible();
		await expect(player.getByAltText('The Echo Location cover art')).toBeVisible();
		await expect(player.getByText('Audio alternative: A course-clear signal')).toBeVisible();
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
		await nonRecipient.getByRole('button', { name: 'Request entry' }).click();
		await expect(nonRecipient.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();
		await page.getByRole('button', { name: /Entry requests, 1 waiting/ }).click();
		const otherRequest = page.getByRole('article').filter({ hasText: 'Other Browser Player' });
		await expect(otherRequest).toBeVisible();
		await otherRequest.getByRole('button', { name: 'Approve' }).click();
		await expect(nonRecipient).toHaveURL(/\/play(?:\/party)?$/);
		await expect.poll(() => nonRecipientAuthorization).not.toBe('');
		await page.goto(gameUrl);

		await page.getByRole('button', { name: 'New announcement' }).first().click();
		announcement = page.getByRole('dialog', { name: 'New announcement' });
		await announcement.getByLabel('Announcement message').fill('Uploaded media announcement');
		await announcement.getByLabel('Recipients').selectOption('player');
		const browserPlayerId = await announcement
			.getByLabel('Player')
			.locator('option')
			.filter({ hasText: /^Seat \d+ · Browser Player$/ })
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
		await announcement.getByLabel('Image description').fill('A one-off status image');
		await announcement
			.getByRole('group', { name: 'Audio (optional)' })
			.getByLabel('Upload for this announcement')
			.check();
		await announcement.getByLabel('Audio file').setInputFiles({
			name: 'one-off.wav',
			mimeType: 'audio/wav',
			buffer: oneSecondWav()
		});
		await announcement.getByLabel('Audio alternative').fill('A one-off status tone');

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
		await expect(page.getByText('The host returned an unexpected response.')).toBeVisible();
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
		await expect(player.getByAltText('A one-off status image')).toBeVisible();
		await expect(player.getByText('Audio alternative: A one-off status tone')).toBeVisible();
		await expect(nonRecipient.getByText('Uploaded media announcement')).not.toBeVisible();
		await expect.poll(() => mediaRequests.length).toBeGreaterThanOrEqual(2);

		const audio = player.locator('audio').last();
		await expect(audio).toHaveAttribute('src', /^blob:/);
		await expect
			.poll(() => audio.evaluate((element) => element.readyState))
			.toBeGreaterThanOrEqual(1);
		const playing = await audio.evaluate(async (element) => {
			element.muted = true;
			element.currentTime = 0;
			await element.play();
			return !element.paused;
		});
		expect(playing).toBe(true);

		const deniedMedia = await nonRecipientContext.request.get(mediaRequests[0].url, {
			headers: { Authorization: nonRecipientAuthorization }
		});
		expect(deniedMedia.status()).toBe(403);
	} finally {
		await playerContext.close();
		await nonRecipientContext.close();
	}
});

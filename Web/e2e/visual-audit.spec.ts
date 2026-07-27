import { expect, test, type Page } from '@playwright/test';
import { mkdirSync } from 'node:fs';
import { resolve } from 'node:path';

const evidenceRoot = resolve(
	import.meta.dirname,
	'..',
	'..',
	'docs',
	'visual-audit',
	'2026-07-27-verified-pass-4'
);
const viewports = [
	{ width: 320, height: 568, key: '320x568' },
	{ width: 390, height: 844, key: '390x844' },
	{ width: 1440, height: 900, key: '1440x900' }
] as const;

const criticalViewports = [
	{ width: 412, height: 915, key: '412x915' },
	{ width: 844, height: 390, key: '844x390' },
	{ width: 768, height: 1024, key: '768x1024' },
	{ width: 1280, height: 720, key: '1280x720' }
] as const;

async function capture(page: Page, area: string, state: string) {
	await captureAtViewports(page, area, state, viewports);
}

async function captureCritical(page: Page, area: string, state: string) {
	await captureAtViewports(page, area, state, criticalViewports);
}

async function captureAtViewports(
	page: Page,
	area: string,
	state: string,
	targets: ReadonlyArray<{ width: number; height: number; key: string }>
) {
	for (const viewport of targets) {
		await page.setViewportSize(viewport);
		await page.screenshot({
			path: resolve(evidenceRoot, `${area}-${state}-${viewport.key}.png`),
			animations: 'disabled'
		});
		const layout = await page.evaluate(() => ({
			pageOverflow: Math.max(0, document.documentElement.scrollWidth - window.innerWidth),
			offscreenControls: Array.from(document.querySelectorAll('button, a, input, select, textarea'))
				.filter((element) => {
					const rect = element.getBoundingClientRect();
					return (
						rect.width > 0 &&
						rect.height > 0 &&
						(rect.left < -1 || rect.right > window.innerWidth + 1)
					);
				})
				.map(
					(element) =>
						element.getAttribute('aria-label') || element.textContent?.trim() || element.tagName
				)
		}));
		expect(layout, `${area}/${state} at ${viewport.key}`).toEqual({
			pageOverflow: 0,
			offscreenControls: []
		});
	}
}

async function captureAtTwoHundredPercent(page: Page, area: string, state: string) {
	await page.setViewportSize({ width: 1440, height: 900 });
	await page.evaluate(() => {
		document.documentElement.style.zoom = '2';
	});
	await page.screenshot({
		path: resolve(evidenceRoot, `${area}-${state}-200-percent.png`),
		animations: 'disabled'
	});
	const layout = await page.evaluate(() => ({
		pageOverflow: Math.max(0, document.documentElement.scrollWidth - window.innerWidth),
		offscreenControls: Array.from(document.querySelectorAll('button, a, input, select, textarea'))
			.filter((element) => {
				const rect = element.getBoundingClientRect();
				return (
					rect.width > 0 &&
					rect.height > 0 &&
					(rect.left < -1 || rect.right > window.innerWidth + 1)
				);
			})
			.map(
				(element) =>
					element.getAttribute('aria-label') || element.textContent?.trim() || element.tagName
			)
	}));
	expect(layout, `${area}/${state} at 200%`).toEqual({
		pageOverflow: 0,
		offscreenControls: []
	});
	await page.evaluate(() => {
		document.documentElement.style.zoom = '';
	});
}

async function clickTask(page: Page, name: string) {
	const navigation = page.getByRole('navigation', { name: 'Game-master dashboard' });
	if (name === 'Approvals') {
		await navigation.getByRole('button', { name: /^Approvals/ }).click();
		return;
	}
	await navigation.getByRole('button', { name, exact: true }).click();
}

test.skip(!process.env.VISUAL_AUDIT, 'Run explicitly with VISUAL_AUDIT=1.');

test('capture the visual overhaul acceptance ledger', async ({ browser, page }) => {
	test.setTimeout(360_000);
	mkdirSync(evidenceRoot, { recursive: true });

	await page.setViewportSize(viewports[0]);
	await page.goto('/');
	await expect(page.getByRole('heading', { name: 'Set up the host' })).toBeVisible();
	await capture(page, 'setup', 'first-launch');

	await page.getByLabel('Username').fill('x');
	await page.getByLabel('Display name').fill('x');
	await page.getByLabel('Password').fill('tiny');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();
	await expect(page.getByText('Please correct the highlighted setup details.')).toBeVisible();
	await capture(page, 'setup', 'validation-error');

	await page.setViewportSize(viewports[2]);
	await page.getByLabel('Username').fill('playwrightadmin');
	await page.getByLabel('Display name').fill('Visual Audit Host');
	await page.getByLabel('Password').fill('secret');
	await page.getByLabel('I understand and trust this local network.').check();
	await page.getByRole('button', { name: 'Create owner' }).click();
	await expect(page.getByRole('button', { name: 'Live', exact: true })).toBeVisible();
	await capture(page, 'host-live', 'no-game');

	const noGameContext = await browser.newContext({ viewport: viewports[0] });
	const noGamePlayer = await noGameContext.newPage();
	await noGamePlayer.goto('/');
	await expect(noGamePlayer.getByRole('heading', { name: 'Enter your profile name' })).toBeVisible();
	await capture(noGamePlayer, 'join', 'no-game');
	await noGamePlayer.goto('/?showQr=1');
	await expect(noGamePlayer.getByRole('heading', { name: 'Scan to join' })).toBeVisible();
	await capture(noGamePlayer, 'join', 'qr');
	await noGameContext.close();

	await clickTask(page, 'Games');
	await expect(page.getByRole('heading', { name: 'Drafts and history' })).toBeVisible();
	await capture(page, 'games', 'empty');
	await page.getByRole('button', { name: 'New game' }).click();
	await expect(page.getByRole('heading', { name: 'New game' })).toBeVisible();
	await capture(page, 'games', 'create-sheet');
	await page.getByLabel('Game name').fill('Visual Audit Round');
	await page.getByLabel('Ruleset').selectOption({ label: 'Blackjack' });
	await page.getByRole('button', { name: 'Create draft' }).click();
	await expect(page.getByRole('heading', { name: 'Visual Audit Round' })).toBeVisible();
	await capture(page, 'host-live', 'draft');

	await page.getByRole('button', { name: 'Open lobby' }).click();
	await expect(page.getByRole('button', { name: 'Close lobby' })).toBeVisible();
	await capture(page, 'host-live', 'lobby-empty');

	const playerOneContext = await browser.newContext({ viewport: viewports[0] });
	const playerOne = await playerOneContext.newPage();
	await playerOne.goto('/');
	await expect(playerOne.getByRole('heading', { name: 'Enter your profile name' })).toBeVisible();
	await capture(playerOne, 'join', 'lobby');
	await playerOne.getByLabel('Profile name').fill('Mira');
	await playerOne.getByRole('button', { name: 'Request entry' }).click();
	await expect(playerOne.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();
	await capture(playerOne, 'join', 'pending');

	const playerTwoContext = await browser.newContext({ viewport: viewports[1] });
	const playerTwo = await playerTwoContext.newPage();
	await playerTwo.goto('/');
	await playerTwo.getByLabel('Profile name').fill('Rowan');
	await playerTwo.getByRole('button', { name: 'Request entry' }).click();
	await expect(playerTwo.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();

	const rejectedContext = await browser.newContext({ viewport: viewports[0] });
	const rejectedPlayer = await rejectedContext.newPage();
	await rejectedPlayer.goto('/');
	await rejectedPlayer.getByLabel('Profile name').fill('Raven');
	await rejectedPlayer.getByRole('button', { name: 'Request entry' }).click();
	await expect(rejectedPlayer.getByRole('heading', { name: 'Awaiting approval' })).toBeVisible();

	await clickTask(page, 'Approvals');
	await expect(page.getByRole('heading', { name: 'Entry requests' })).toBeVisible();
	await capture(page, 'approvals', 'pending');
	const rejectedRequest = page.locator('article').filter({ hasText: 'Raven' });
	await rejectedRequest.getByRole('button', { name: 'Reject' }).click();
	await expect(page.getByRole('heading', { name: 'Reject profile request' })).toBeVisible();
	await capture(page, 'approvals', 'rejection-sheet');
	await page.getByLabel('Reason').fill('That name is reserved for another player.');
	await page.getByRole('button', { name: 'Reject request' }).click();
	await expect(rejectedPlayer.getByRole('heading', { name: 'Request update' })).toBeVisible();
	await capture(rejectedPlayer, 'join', 'rejected');
	for (const name of ['Mira', 'Rowan']) {
		const requestCard = page.locator('article').filter({ hasText: name });
		await requestCard.getByRole('button', { name: 'Approve' }).click();
	}
	await expect(page.getByRole('heading', { name: 'No pending requests' })).toBeVisible();
	await capture(page, 'approvals', 'empty');

	await expect(playerOne).toHaveURL(/\/play$/);
	await expect(playerTwo).toHaveURL(/\/play$/);
	await capture(playerOne, 'player-stage', 'lobby');

	await page.reload();
	await expect(page.getByRole('heading', { name: 'Visual Audit Round' })).toBeVisible();
	await clickTask(page, 'Live');
	await expect(page.getByRole('combobox', { name: 'Role for Mira' })).toBeVisible();
	await page.getByRole('button', { name: 'Randomize' }).click();
	await page.getByRole('button', { name: 'Save roles' }).click();
	await page.getByRole('button', { name: 'Start game' }).click();
	await expect(page.getByRole('button', { name: 'Pause game' })).toBeVisible();
	await page.getByRole('button', { name: 'Start', exact: true }).click();
	await capture(page, 'host-live', 'running');
	await captureCritical(page, 'host-live', 'running-critical');

	await playerOne.reload();
	await expect(playerOne.getByRole('button', { name: 'Role', exact: true })).toBeVisible();
	await capture(playerOne, 'player-stage', 'running-timer');
	await captureCritical(playerOne, 'player-stage', 'running-critical');
	await playerOne.getByRole('button', { name: 'Role', exact: true }).click();
	await expect(playerOne.getByRole('button', { name: 'Reveal', exact: true })).toBeVisible();
	await capture(playerOne, 'role', 'concealed');
	await playerOne.getByRole('button', { name: 'Reveal', exact: true }).click();
	await capture(playerOne, 'role', 'revealed');
	await playerOne.getByRole('button', { name: 'Close Your role' }).click();

	await playerOne.getByRole('button', { name: 'More', exact: true }).click();
	await page
		.getByPlaceholder('Announcement to every player…')
		.fill('Meet beside the old oak after the bell. Bring your role card and wait for the host.');
	await page.getByRole('button', { name: 'Announce' }).click();
	await capture(playerOne, 'attention', 'pending-behind-sheet');
	await playerOne.getByRole('button', { name: 'Close More' }).click();
	await expect(playerOne.getByRole('article', { name: 'Announcement 1 of 1' })).toBeVisible();
	await capture(playerOne, 'attention', 'presented-after-sheet');
	await playerOne.reload();
	await expect(playerOne.getByRole('article', { name: 'Announcement 1 of 1' })).toBeVisible();
	await capture(playerOne, 'attention', 'reload-persistence');
	await playerOne.getByRole('button', { name: 'Acknowledge' }).click();

	await playerOne.getByRole('button', { name: 'Open chat' }).click();
	await capture(playerOne, 'chat', 'channel-list');
	await captureCritical(playerOne, 'chat', 'channel-list-critical');
	await playerOne.getByRole('button', { name: 'General' }).click();
	await expect(playerOne.getByPlaceholder(/Write a message/)).toBeVisible();
	await capture(playerOne, 'chat', 'active-conversation');
	await playerOne.getByPlaceholder(/Write a message/).fill('Ready at the table.');
	await playerOne.getByRole('button', { name: 'Send' }).click();
	await expect(playerOne.getByText('Ready at the table.')).toBeVisible();
	await capture(playerOne, 'chat', 'message-sent');
	await playerOne.getByRole('button', { name: 'Close General' }).click();
	await playerOne.getByRole('button', { name: 'Party', exact: true }).click();
	await capture(playerOne, 'party', 'roster-sheet');
	await playerOne.getByRole('button', { name: /Rowan/ }).click();
	await expect(playerOne.getByRole('heading', { name: 'Rowan' })).toBeVisible();
	await capture(playerOne, 'party', 'member-profile');
	await playerOne.getByRole('button', { name: 'Close Party' }).click();
	await playerOne.getByRole('button', { name: 'More', exact: true }).click();
	await capture(playerOne, 'more', 'settings');
	await playerOne.getByRole('button', { name: 'Profile' }).click();
	await capture(playerOne, 'more', 'profile');
	await playerOne.locator('.back-link').click();
	await playerOne.getByRole('button', { name: 'History' }).click();
	await expect(playerOne.getByText('No completed games yet.')).toBeVisible();
	await capture(playerOne, 'more', 'history-empty');
	await playerOne.locator('.back-link').click();
	await playerOne.getByRole('button', { name: 'Connection details' }).click();
	await capture(playerOne, 'more', 'connection');
	await playerOne.locator('.back-link').click();
	await playerOne.getByRole('button', { name: 'Display' }).click();
	await playerOne.getByLabel('Large text').check();
	await playerOne.getByLabel('High contrast').check();
	await capture(playerOne, 'display-modes', 'large-high-contrast');
	await captureAtTwoHundredPercent(playerOne, 'display-modes', 'large-high-contrast');
	await playerOne.getByRole('button', { name: 'Close More' }).click();

	await page.getByRole('button', { name: 'Pause game' }).click();
	await capture(page, 'host-live', 'paused');
	await playerOne.reload();
	await expect(playerOne.getByRole('heading', { name: 'Waiting for phase' })).toBeVisible();
	await capture(playerOne, 'player-stage', 'paused');

	const lifecycleControls = page.locator('.command-bar');
	await lifecycleControls.getByRole('button', { name: 'Resume', exact: true }).click();
	await lifecycleControls.getByRole('button', { name: 'Begin review' }).click();
	await capture(page, 'host-live', 'review');
	await clickTask(page, 'Games');
	await capture(page, 'games', 'populated');
	await clickTask(page, 'Live');
	await page.getByRole('button', { name: 'Archive', exact: true }).click();
	await expect(page.locator('.game-title .muted').filter({ hasText: 'archived' })).toBeVisible();
	await capture(page, 'host-live', 'archived');
	await playerOne.reload();
	await capture(playerOne, 'player-stage', 'completed');

	await clickTask(page, 'Games');
	await capture(page, 'games', 'after-archive-empty');
	await page.getByLabel('Show archived games').check();
	await capture(page, 'games', 'archived');

	await clickTask(page, 'More');
	await page.getByRole('button', { name: /^Rulesets/ }).click();
	await expect(page.getByRole('heading', { name: 'Ruleset library' })).toBeVisible();
	await capture(page, 'rulesets', 'populated');
	await page.getByRole('link', { name: /^Town of Salem/ }).click();
	await expect(page.getByRole('heading', { name: 'Overview and limits' })).toBeVisible();
	await capture(page, 'ruleset-editor', 'overview');
	await captureCritical(page, 'ruleset-editor', 'overview-critical');
	await page.setViewportSize(viewports[0]);
	await page.getByRole('button', { name: 'Sections' }).click();
	await capture(page, 'ruleset-editor', 'sections-sheet');
	await page.getByRole('button', { name: 'Close Ruleset sections' }).click();

	await page.goto('/admin');
	await page.waitForTimeout(500);
	await clickTask(page, 'More');
	await expect(page.getByRole('heading', { name: 'More' })).toBeVisible();
	await page.waitForTimeout(250);
	await page.getByRole('button', { name: /^Installation/ }).click();
	await expect(page.getByRole('heading', { name: 'Network hosting' })).toBeVisible();
	await capture(page, 'installation', 'network');
	for (const task of ['Phone join', 'Game masters', 'Backups', 'Diagnostics']) {
		await page.getByRole('button', { name: task, exact: true }).click();
		await capture(page, 'installation', task.toLowerCase().replace(' ', '-'));
	}

	await playerOneContext.close();
	await playerTwoContext.close();
	await rejectedContext.close();
});

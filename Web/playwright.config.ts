import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './e2e',
	fullyParallel: false,
	workers: 1,
	retries: 0,
	timeout: 30_000,
	reporter: [['list'], ['html', { open: 'never' }]],
	globalSetup: './e2e/build-host.mjs',
	use: {
		trace: 'retain-on-failure'
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
});

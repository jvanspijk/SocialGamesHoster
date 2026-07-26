import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
	testDir: './e2e',
	fullyParallel: false,
	workers: 1,
	retries: 0,
	timeout: 30_000,
	reporter: [['list'], ['html', { open: 'never' }]],
	use: {
		baseURL: 'http://127.0.0.1:19090',
		trace: 'retain-on-failure'
	},
	webServer: {
		command: 'node e2e/start-host.mjs',
		url: 'http://127.0.0.1:19090/api/app/v1/setup/status',
		reuseExistingServer: false,
		timeout: 120_000
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }]
});

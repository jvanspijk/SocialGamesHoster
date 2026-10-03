import { spawn } from 'node:child_process';
import { mkdtemp, mkdir, rm } from 'node:fs/promises';
import { createServer } from 'node:net';
import { resolve } from 'node:path';
import { setTimeout as delay } from 'node:timers/promises';
import { test as base, expect } from '@playwright/test';

const webRoot = resolve(import.meta.dirname, '..');
const projectRoot = resolve(webRoot, '..');

async function availablePort() {
	const server = createServer();
	await new Promise<void>((resolve, reject) => {
		server.once('error', reject);
		server.listen(0, '127.0.0.1', resolve);
	});
	const address = server.address();
	if (!address || typeof address === 'string') throw new Error('Cannot allocate an E2E port');
	await new Promise<void>((resolve, reject) =>
		server.close((error) => (error ? reject(error) : resolve()))
	);
	return address.port;
}

// The browser context is already per test; give it an equally isolated host and database.
export const test = base.extend({
	baseURL: [
		// Playwright requires a destructured fixture parameter, even without dependencies.
		// eslint-disable-next-line no-empty-pattern
		async ({}, use) => {
			const dataRoot = resolve(webRoot, '.e2e-data');
			await mkdir(dataRoot, { recursive: true });
			const dataDir = await mkdtemp(resolve(dataRoot, 'test-'));
			const port = await availablePort();
			const baseURL = `http://127.0.0.1:${port}`;
			const binary = resolve(
				webRoot,
				'.test-bin',
				process.platform === 'win32' ? 'e2e-host.exe' : 'e2e-host'
			);
			const host = spawn(binary, ['--no-tray', `--http=127.0.0.1:${port}`, `--dir=${dataDir}`], {
				cwd: projectRoot,
				stdio: ['ignore', 'pipe', 'pipe']
			});
			let output = '';
			let failure: Error | undefined;
			host.stdout.on('data', (chunk) => (output += chunk.toString()));
			host.stderr.on('data', (chunk) => (output += chunk.toString()));
			host.on('error', (error) => (failure = error));
			const stopped = new Promise<void>((resolve) => host.once('close', () => resolve()));
			try {
				const deadline = Date.now() + 60_000;
				while (true) {
					if (failure) throw failure;
					if (host.exitCode !== null || host.signalCode !== null)
						throw new Error(`E2E host exited before becoming ready:\n${output}`);
					let ready = false;
					try {
						const response = await fetch(`${baseURL}/api/app/v1/setup/status`, {
							signal: AbortSignal.timeout(1_000)
						});
						ready = response.ok && (await response.json()).needsOwner === true;
					} catch {
						// The host may still be applying migrations or importing its bundled fixtures.
					}
					if (ready) break;
					if (Date.now() >= deadline) throw new Error(`E2E host startup timed out:\n${output}`);
					await delay(100);
				}
				await use(baseURL);
			} finally {
				if (host.exitCode === null && host.signalCode === null) host.kill('SIGTERM');
				await Promise.race([stopped, delay(5_000, undefined, { ref: false })]);
				if (host.exitCode === null && host.signalCode === null) host.kill('SIGKILL');
				await stopped;
				await rm(dataDir, { recursive: true, force: true });
			}
		},
		{ timeout: 120_000 }
	]
});

export { expect };

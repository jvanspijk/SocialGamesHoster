import { spawn } from 'node:child_process';
import { resolve } from 'node:path';

const webRoot = resolve(import.meta.dirname, '..');
const cli = resolve(webRoot, 'node_modules', '@playwright', 'test', 'cli.js');
const child = spawn(process.execPath, [cli, 'test'], {
	cwd: webRoot,
	stdio: 'inherit',
	env: {
		...process.env,
		PLAYWRIGHT_BROWSERS_PATH: resolve(webRoot, '.playwright-browsers')
	}
});

child.on('exit', (code) => process.exit(code ?? 1));

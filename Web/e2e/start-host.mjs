import { existsSync, mkdirSync, rmSync } from 'node:fs';
import { spawn, spawnSync } from 'node:child_process';
import { resolve } from 'node:path';

const webRoot = resolve(import.meta.dirname, '..');
const projectRoot = resolve(webRoot, '..');
const dataDir = resolve(webRoot, '.e2e-data');
const localGo = resolve(
	projectRoot,
	'.tools',
	'go',
	'bin',
	process.platform === 'win32' ? 'go.exe' : 'go'
);
const go = existsSync(localGo) ? localGo : 'go';
const binaryDir = resolve(webRoot, '.test-bin');
const hostBinary = resolve(binaryDir, process.platform === 'win32' ? 'e2e-host.exe' : 'e2e-host');
const goEnvironment = {
	...process.env,
	CGO_ENABLED: '0'
};

rmSync(dataDir, { recursive: true, force: true });
mkdirSync(binaryDir, { recursive: true });
const build = spawnSync(go, ['build', '-o', hostBinary, './Host/cmd/socialgameshoster'], {
	cwd: projectRoot,
	stdio: 'inherit',
	env: goEnvironment
});
if (build.status !== 0) process.exit(build.status ?? 1);

const child = spawn(hostBinary, ['--no-tray', '--http=127.0.0.1:19090', `--dir=${dataDir}`], {
	cwd: projectRoot,
	stdio: 'inherit',
	env: goEnvironment
});

const stop = () => {
	if (!child.killed) child.kill('SIGTERM');
};
process.on('SIGTERM', stop);
process.on('SIGINT', stop);
child.on('exit', (code) => process.exit(code ?? 0));

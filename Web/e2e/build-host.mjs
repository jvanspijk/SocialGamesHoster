import { existsSync, mkdirSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { resolve } from 'node:path';

const webRoot = resolve(import.meta.dirname, '..');
const projectRoot = resolve(webRoot, '..');
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

export default function buildHost() {
	mkdirSync(binaryDir, { recursive: true });
	const build = spawnSync(go, ['build', '-o', hostBinary, './Host/cmd/socialgameshoster'], {
		cwd: projectRoot,
		stdio: 'inherit',
		env: goEnvironment
	});
	if (build.error) throw build.error;
	if (build.status !== 0) throw new Error(`E2E host build failed with exit code ${build.status}`);
}

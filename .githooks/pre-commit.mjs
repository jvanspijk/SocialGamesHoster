import { spawnSync } from 'node:child_process';
import { join } from 'node:path';

const repoRoot = runGit(['rev-parse', '--show-toplevel']).trim();
const stagedOutput = runGit([
	'diff',
	'--cached',
	'--name-only',
	'--diff-filter=ACMR',
	'-z',
	'--',
	'Web',
]);
const stagedWebFiles = stagedOutput
	.split('\0')
	.filter((file) => file.length > 0 && file.startsWith('Web/'));

if (stagedWebFiles.length === 0) {
	process.exit(0);
}

const npmCommand = process.platform === 'win32' ? 'npm.cmd' : 'npm';
const relativeWebFiles = stagedWebFiles.map((file) => file.slice('Web/'.length));
const formatResult = spawnSync(
	npmCommand,
	['run', 'format:staged', '--', ...relativeWebFiles],
	{
		cwd: join(repoRoot, 'Web'),
		stdio: 'inherit',
		shell: process.platform === 'win32',
	},
);

if (formatResult.error) {
	console.error(`Unable to run the frontend formatter: ${formatResult.error.message}`);
	process.exit(1);
}

if (formatResult.status !== 0) {
	process.exit(formatResult.status ?? 1);
}

const stageResult = spawnSync('git', ['add', '--', ...stagedWebFiles], {
	cwd: repoRoot,
	stdio: 'inherit',
});

if (stageResult.error) {
	console.error(`Unable to stage formatted frontend files: ${stageResult.error.message}`);
	process.exit(1);
}

process.exit(stageResult.status ?? 1);

function runGit(arguments_) {
	const result = spawnSync('git', arguments_, {
		cwd: process.cwd(),
		encoding: 'utf8',
	});

	if (result.error) {
		console.error(`Unable to inspect staged files: ${result.error.message}`);
		process.exit(1);
	}

	if (result.status !== 0) {
		process.exit(result.status ?? 1);
	}

	return result.stdout;
}

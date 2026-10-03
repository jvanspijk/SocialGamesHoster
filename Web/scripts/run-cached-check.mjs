import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { spawnSync } from 'node:child_process';
import { resolve } from 'node:path';

const webRoot = resolve(import.meta.dirname, '..');
// Tool caches account for source contents and check configuration. Include the
// manifests too so upgrading a parser or plugin cannot reuse its old results.
const fingerprint = createHash('sha256')
	.update(readFileSync(resolve(webRoot, 'package.json')))
	.update(readFileSync(resolve(webRoot, 'package-lock.json')))
	.digest('hex');
const cacheRoot = resolve(webRoot, 'node_modules', '.cache', 'sgh-checks', fingerprint);
const checks = {
	prettier: [
		'node_modules/prettier/bin/prettier.cjs',
		'--check',
		'.',
		'--cache',
		'--cache-strategy',
		'content',
		'--cache-location',
		resolve(cacheRoot, 'prettier')
	],
	eslint: [
		'node_modules/eslint/bin/eslint.js',
		'.',
		'--cache',
		'--cache-strategy',
		'content',
		'--cache-location',
		resolve(cacheRoot, 'eslint')
	]
};
const args = checks[process.argv[2]];
if (!args) throw new Error('Expected prettier or eslint.');
const result = spawnSync(process.execPath, args, { cwd: webRoot, stdio: 'inherit' });
if (result.error) throw result.error;
process.exit(result.status ?? 1);

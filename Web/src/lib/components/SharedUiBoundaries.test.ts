import { readdirSync, readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { describe, expect, it } from 'vitest';

const directory = resolve(process.cwd(), 'src/lib/components');
const prohibitedImport = /\$lib\/(?:api|state|forms)(?:\/|['"])/;

describe('shared UI boundaries', () => {
	it('recognizes both quote styles for prohibited application imports', () => {
		expect(`import { connection } from "$lib/state/connection.svelte";`).toMatch(prohibitedImport);
		expect(`import { api } from '$lib/api/client';`).toMatch(prohibitedImport);
	});

	it('does not import application transport, stores, or form error models', () => {
		for (const file of readdirSync(directory).filter((entry) => entry.endsWith('.svelte'))) {
			const source = readFileSync(`${directory}/${file}`, 'utf8');
			expect(source).not.toMatch(prohibitedImport);
		}
	});
});

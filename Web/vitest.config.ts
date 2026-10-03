import { sveltekit } from '@sveltejs/kit/vite';
import { configDefaults, defineConfig } from 'vitest/config';

// Browser-free suites must not load the DOM matchers, storage, or dialog setup.
// Keep other suites in jsdom unless their imports and assertions work in Node.
const nodeTests = [
	'src/lib/api/errors.test.ts',
	'src/lib/forms/errors.test.ts',
	'src/lib/state/chatReadMarkers.test.ts',
	'src/lib/features/games/roleAssignments.test.ts',
	'src/lib/features/media/media.test.ts',
	'src/lib/features/rulesets/editor-state.test.ts',
	'src/lib/features/rulesets/components/definition-editor.test.ts',
	'src/lib/components/SharedUiBoundaries.test.ts'
];

export default defineConfig({
	plugins: [sveltekit()],
	resolve: {
		conditions: ['browser']
	},
	test: {
		clearMocks: true,
		projects: [
			{
				extends: true,
				test: {
					name: 'node',
					environment: 'node',
					include: nodeTests
				}
			},
			{
				extends: true,
				test: {
					name: 'dom',
					environment: 'jsdom',
					setupFiles: ['./src/test/setup.ts'],
					include: ['src/**/*.test.ts'],
					exclude: [...configDefaults.exclude, ...nodeTests]
				}
			}
		]
	}
});

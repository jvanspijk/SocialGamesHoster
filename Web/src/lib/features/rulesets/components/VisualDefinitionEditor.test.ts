import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import VisualDefinitionEditorHarness from '../../../../test/VisualDefinitionEditorHarness.svelte';

describe('VisualDefinitionEditor', () => {
	afterEach(cleanup);

	it('renders focus targets that match nested issue destinations', () => {
		render(VisualDefinitionEditorHarness, { props: { section: 'knowledge' } });
		expect(document.querySelector('#field-knowledge-viewer-0')).toBeInTheDocument();
		expect(document.querySelector('[name="knowledge-viewer-0-tags"]')).toBeInTheDocument();
		expect(document.querySelector('[name="knowledge-reveal-0"]')).toBeInTheDocument();
	});

	it.each([
		['teams', 'Teams'],
		['roles', 'Abilities'],
		['phases', 'Phases'],
		['knowledge', 'Starting knowledge'],
		['chat', 'Default chat settings'],
		['achievements', 'Achievements'],
		['assets', 'Sounds']
	] as const)('dispatches the %s section to its feature-local editor', (section, heading) => {
		render(VisualDefinitionEditorHarness, { props: { section } });

		expect(screen.getByRole('heading', { name: heading })).toBeVisible();
	});
});

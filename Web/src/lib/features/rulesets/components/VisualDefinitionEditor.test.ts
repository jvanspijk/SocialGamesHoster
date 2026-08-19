import { cleanup, fireEvent, render, screen, within } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import VisualDefinitionEditorHarness from '../../../../test/VisualDefinitionEditorHarness.svelte';

describe('VisualDefinitionEditor', () => {
	afterEach(cleanup);

	it('uses dense content headers for nested composition titles and actions', async () => {
		render(VisualDefinitionEditorHarness);

		const roleSlots = screen.getByText('Role slots').closest('.content-header');
		const slotChanges = screen.getByText('Slot changes').closest('.content-header');

		expect(roleSlots).toHaveClass('dense', 'has-actions');
		expect(slotChanges).toHaveClass('dense', 'has-actions');
		await fireEvent.click(screen.getByRole('button', { name: 'Add slot' }));
		expect(screen.getByText('Role slot', { selector: 'strong' })).toBeVisible();
	});

	it('navigates from composition reference warnings to the affected modifier', async () => {
		const onnavigate = vi.fn();
		render(VisualDefinitionEditorHarness, { props: { onnavigate } });

		const inlineWarning = screen.getByText(/^Used by/);
		await fireEvent.click(
			within(inlineWarning).getByRole('button', { name: 'Conditional change 1' })
		);
		expect(onnavigate).toHaveBeenLastCalledWith('composition', 'modifier-1');

		await fireEvent.click(screen.getByRole('button', { name: 'Delete 3–8 players' }));
		await fireEvent.click(
			within(screen.getByRole('dialog')).getByRole('button', { name: 'Conditional change 1' })
		);
		expect(onnavigate).toHaveBeenLastCalledWith('composition', 'modifier-1');
	});

	it('renders focus targets that match nested issue destinations', () => {
		render(VisualDefinitionEditorHarness);
		expect(document.querySelector('[name="slot-count-0-0"]')).toBeInTheDocument();
		expect(document.querySelector('#field-slot-selector-0-0')).toBeInTheDocument();
		cleanup();

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
		['chat', 'Normal chat settings'],
		['achievements', 'Achievements'],
		['audio', 'Audio cues']
	] as const)('dispatches the %s section to its feature-local editor', (section, heading) => {
		render(VisualDefinitionEditorHarness, { props: { section } });

		expect(screen.getByRole('heading', { name: heading })).toBeVisible();
	});
});

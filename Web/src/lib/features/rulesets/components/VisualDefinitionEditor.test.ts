import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
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
		expect(screen.getByText('Role slot')).toBeVisible();
	});
});

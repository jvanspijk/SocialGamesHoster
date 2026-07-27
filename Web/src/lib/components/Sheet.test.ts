import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import SheetHarness from '../../test/SheetHarness.svelte';

describe('Sheet', () => {
	it('names the modal, moves focus inside, and restores the trigger on close', async () => {
		render(SheetHarness);
		const trigger = screen.getByRole('button', { name: 'Open settings' });

		trigger.focus();
		await fireEvent.click(trigger);
		const dialog = screen.getByRole('dialog', { name: 'Display settings' });
		expect(dialog).toHaveAttribute('open');
		await Promise.resolve();
		expect(screen.getByRole('heading', { name: 'Display settings' })).toHaveFocus();

		await fireEvent.click(screen.getByRole('button', { name: 'Close Display settings' }));
		expect(dialog).not.toHaveAttribute('open');
		await waitFor(() => expect(trigger).toHaveFocus());
	});
});

import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import SheetHarness from '../../test/SheetHarness.svelte';

function finishOverlayExit(target: HTMLElement) {
	const event = new Event('transitionend', { bubbles: true });
	Object.defineProperty(event, 'propertyName', { value: 'opacity' });
	return fireEvent(target, event);
}

describe('Sheet', () => {
	afterEach(cleanup);

	it('uses the sheet presentation and keeps nested updates open', async () => {
		render(SheetHarness);
		const trigger = screen.getByRole('button', { name: 'Open settings' });

		trigger.focus();
		await fireEvent.click(trigger);
		const dialog = screen.getByRole('dialog', { name: 'Display settings' });
		expect(dialog).toHaveAttribute('data-presentation', 'sheet');
		expect(dialog).toHaveAttribute('open');

		await fireEvent.click(screen.getByRole('button', { name: 'Update sheet content' }));
		expect(screen.getByTestId('updates')).toHaveTextContent('1');
		expect(dialog).toHaveAttribute('open');
	});

	it('starts its tokenized exit once on Escape before native close', async () => {
		render(SheetHarness);
		const trigger = screen.getByRole('button', { name: 'Open settings' });
		trigger.focus();
		await fireEvent.click(trigger);
		const dialog = screen.getByRole('dialog', { name: 'Display settings' });

		await fireEvent(dialog, new Event('cancel', { cancelable: true }));
		expect(dialog).toHaveAttribute('open');
		expect(dialog).toHaveClass('closing');
		await finishOverlayExit(dialog);
		expect(dialog).not.toHaveAttribute('open');
		expect(screen.getByTestId('close-count')).toHaveTextContent('1');
	});

	it('synchronizes explicit and external closes without another close callback', async () => {
		render(SheetHarness);
		const trigger = screen.getByRole('button', { name: 'Open settings' });

		trigger.focus();
		await fireEvent.click(trigger);
		await fireEvent.click(screen.getByRole('button', { name: 'Close Display settings' }));
		const dialog = screen.getByRole('dialog', { name: 'Display settings' });
		expect(dialog).toHaveClass('closing');
		await finishOverlayExit(dialog);
		expect(screen.getByTestId('close-count')).toHaveTextContent('1');

		await fireEvent.click(trigger);
		await fireEvent.click(screen.getByRole('button', { name: 'Close externally' }));
		expect(dialog).toHaveAttribute('open');
		expect(dialog).toHaveClass('closing');
		await finishOverlayExit(dialog);
		expect(dialog).not.toHaveAttribute('open');
		expect(screen.getByTestId('close-count')).toHaveTextContent('1');
	});
});

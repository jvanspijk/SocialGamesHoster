import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import DialogHarness from '../../test/DialogHarness.svelte';

function finishOverlayExit(target: HTMLElement) {
	const event = new Event('transitionend', { bubbles: true });
	Object.defineProperty(event, 'propertyName', { value: 'opacity' });
	return fireEvent(target, event);
}

describe('Dialog', () => {
	afterEach(cleanup);

	it('opens with an accessible name and keeps nested updates open', async () => {
		render(DialogHarness);
		const trigger = screen.getByRole('button', { name: 'Open recipients' });

		trigger.focus();
		await fireEvent.click(trigger);
		const dialog = screen.getByRole('dialog', { name: 'New message' });
		expect(dialog).toHaveAttribute('aria-describedby');
		expect(dialog).toHaveAttribute('open');

		await fireEvent.click(screen.getByRole('button', { name: 'Update dialog content' }));
		expect(screen.getByTestId('updates')).toHaveTextContent('1');
		expect(dialog).toHaveAttribute('open');

		const recipient = screen.getByRole('button', { name: /Rowan/ });
		expect(recipient).toBeVisible();
		await fireEvent.click(recipient);
		expect(screen.getByTestId('chosen')).toHaveTextContent('Rowan');
	});

	it('starts its tokenized exit once on Escape before native close', async () => {
		render(DialogHarness);
		const trigger = screen.getByRole('button', { name: 'Open recipients' });
		trigger.focus();
		await fireEvent.click(trigger);
		const dialog = screen.getByRole('dialog', { name: 'New message' });

		await fireEvent(dialog, new Event('cancel', { cancelable: true }));
		expect(dialog).toHaveAttribute('open');
		expect(dialog).toHaveClass('closing');
		await finishOverlayExit(dialog);
		expect(dialog).not.toHaveAttribute('open');
		expect(screen.getByTestId('close-count')).toHaveTextContent('1');
	});

	it('synchronizes explicit and external closes without another close callback', async () => {
		render(DialogHarness);
		const trigger = screen.getByRole('button', { name: 'Open recipients' });

		trigger.focus();
		await fireEvent.click(trigger);
		await fireEvent.click(screen.getByRole('button', { name: 'Close New message' }));
		const dialog = screen.getByRole('dialog', { name: 'New message' });
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

	it('ignores a bubbled child transition while its exit is pending', async () => {
		render(DialogHarness);
		await fireEvent.click(screen.getByRole('button', { name: 'Open recipients' }));
		const dialog = screen.getByRole('dialog', { name: 'New message' });

		await fireEvent.click(screen.getByRole('button', { name: 'Close New message' }));
		await finishOverlayExit(screen.getByTestId('updates'));
		expect(dialog).toHaveAttribute('open');
		expect(dialog).toHaveClass('closing');

		await finishOverlayExit(dialog);
		expect(dialog).not.toHaveAttribute('open');
	});
});

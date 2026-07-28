import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import DialogHarness from '../../test/DialogHarness.svelte';

describe('Dialog', () => {
	it('keeps recipient controls tappable inside the modal', async () => {
		render(DialogHarness);

		await fireEvent.click(screen.getByRole('button', { name: 'Open recipients' }));
		const recipient = screen.getByRole('button', { name: /Rowan/ });
		expect(recipient).toBeVisible();

		await fireEvent.click(recipient);
		expect(screen.getByTestId('chosen')).toHaveTextContent('Rowan');
	});
});

import { createRawSnippet } from 'svelte';
import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import IconButton from './IconButton.svelte';

const icon = createRawSnippet(() => ({ render: () => '<svg aria-hidden="true"></svg>' }));

describe('IconButton', () => {
	it('requires a name, invokes its callback, and receives keyboard focus', async () => {
		const onclick = vi.fn();
		render(IconButton, {
			props: { label: 'Close panel', icon, onclick }
		});

		const button = screen.getByRole('button', { name: 'Close panel' });
		button.focus();
		expect(button).toHaveFocus();

		await fireEvent.click(button);
		expect(onclick).toHaveBeenCalledOnce();
	});

	it('disables while loading and exposes its loading status', () => {
		render(IconButton, {
			props: { label: 'Save changes', icon, loading: true }
		});

		const button = screen.getByRole('button', { name: 'Loading Save changes' });
		expect(button).toBeDisabled();
		expect(button).toHaveAttribute('aria-busy', 'true');
	});

	it('supports disabled and danger states', () => {
		render(IconButton, {
			props: { label: 'Delete game', icon, disabled: true, variant: 'danger' }
		});

		const button = screen.getByRole('button', { name: 'Delete game' });
		expect(button).toBeDisabled();
		expect(button).toHaveAttribute('data-variant', 'danger');
	});
});

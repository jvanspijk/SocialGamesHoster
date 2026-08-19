import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it } from 'vitest';
import CollectionEditorHarness from '../../../../test/CollectionEditorHarness.svelte';

describe('CollectionEditor', () => {
	afterEach(cleanup);
	it('supports keyboard selection, duplicate, reorder, and delete', async () => {
		const user = userEvent.setup();
		render(CollectionEditorHarness);
		const one = screen.getByRole('button', { name: 'One' });
		one.focus();
		await fireEvent.keyDown(one, { key: 'ArrowDown' });
		expect(screen.getByRole('button', { name: 'Two' })).toHaveAttribute('aria-current', 'page');
		await user.click(screen.getByRole('button', { name: 'Duplicate Two' }));
		expect(screen.getByRole('button', { name: 'Two copy' })).toBeVisible();
		await user.click(screen.getByRole('button', { name: 'Move Two down' }));
		await user.click(screen.getByRole('button', { name: 'Delete Two' }));
		await user.click(screen.getByRole('button', { name: /^Delete$/ }));
		expect(screen.queryByRole('button', { name: 'Two' })).not.toBeInTheDocument();
	});

	it('blocks deletion and names incoming references', async () => {
		const user = userEvent.setup();
		render(CollectionEditorHarness);
		await user.click(screen.getByRole('button', { name: 'Delete One' }));
		expect(screen.getByText('Role · Villager')).toBeVisible();
		expect(screen.queryByRole('button', { name: /^Delete$/ })).not.toBeInTheDocument();
	});
});

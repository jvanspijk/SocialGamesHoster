import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import SelectionDialog from './SelectionDialog.svelte';

function finishOverlayExit(target: HTMLElement) {
	const event = new Event('transitionend', { bubbles: true });
	Object.defineProperty(event, 'propertyName', { value: 'opacity' });
	return fireEvent(target, event);
}

const entries = [
	{
		id: 'participant-2',
		label: 'Rowan',
		accessibleLabel: 'Rowan, Seat 2',
		supportingLabel: 'Seat 2',
		leadingText: 'R'
	}
];

afterEach(cleanup);

describe('SelectionDialog', () => {
	it('labels the dialog and exposes supporting content on selectable buttons', async () => {
		const onselect = vi.fn();
		render(SelectionDialog, {
			props: {
				open: true,
				title: 'Choose a destination',
				description: 'Select where to continue.',
				entries,
				emptyState: { title: 'Nothing available', description: 'Try again later.' },
				onselect,
				close: vi.fn()
			}
		});

		const dialog = screen.getByRole('dialog', { name: 'Choose a destination' });
		expect(dialog).toHaveAttribute('open');
		const descriptionId = dialog.getAttribute('aria-describedby');
		expect(descriptionId).toBeTruthy();
		expect(document.getElementById(descriptionId ?? '')).toHaveTextContent(
			'Select where to continue.'
		);

		const entry = screen.getByRole('button', { name: 'Rowan, Seat 2' });
		expect(entry).toBeInstanceOf(HTMLButtonElement);
		expect(entry).toHaveTextContent('R');
		expect(entry).toHaveTextContent('Rowan');
		expect(entry).toHaveTextContent('Seat 2');
		entry.focus();
		await fireEvent.click(entry);
		expect(onselect).toHaveBeenCalledWith('participant-2');
	});

	it.each([
		['Enter', '{Enter}'],
		['Space', ' ']
	])('activates entries with %s through native button semantics', async (_key, input) => {
		const onselect = vi.fn();
		const user = userEvent.setup();
		render(SelectionDialog, {
			props: {
				open: true,
				title: 'Choose a destination',
				description: 'Select where to continue.',
				entries,
				emptyState: { title: 'Nothing available', description: 'Try again later.' },
				onselect,
				close: vi.fn()
			}
		});

		const entry = screen.getByRole('button', { name: 'Rowan, Seat 2' });
		entry.focus();
		await user.keyboard(input);
		expect(onselect).toHaveBeenCalledWith('participant-2');
	});

	it('renders caller-owned empty-state content', () => {
		render(SelectionDialog, {
			props: {
				open: true,
				title: 'Choose a destination',
				description: 'Select where to continue.',
				entries: [],
				emptyState: {
					title: 'Nothing available',
					description: 'There are no destinations to choose.'
				},
				onselect: vi.fn(),
				close: vi.fn()
			}
		});

		expect(screen.getByRole('heading', { name: 'Nothing available' })).toBeVisible();
		expect(screen.getByText('There are no destinations to choose.')).toBeVisible();
		expect(screen.queryByRole('button', { name: 'Rowan, Seat 2' })).not.toBeInTheDocument();
	});

	it('closes explicitly and restores focus to the trigger', async () => {
		const trigger = document.createElement('button');
		trigger.type = 'button';
		trigger.textContent = 'Open choices';
		document.body.append(trigger);
		trigger.focus();

		try {
			let rerender: (props: { open: boolean }) => Promise<void> = async () => undefined;
			const rendered = render(SelectionDialog, {
				props: {
					open: true,
					title: 'Choose a destination',
					description: 'Select where to continue.',
					entries,
					emptyState: { title: 'Nothing available', description: 'Try again later.' },
					onselect: vi.fn(),
					close: () => void rerender({ open: false })
				}
			});
			rerender = rendered.rerender;

			const dialog = screen.getByRole('dialog', { name: 'Choose a destination' });
			const close = screen.getByRole('button', { name: 'Close Choose a destination' });
			close.focus();
			await fireEvent.click(close);
			await waitFor(() => expect(dialog).toHaveClass('closing'));
			await finishOverlayExit(dialog);
			expect(dialog).not.toHaveAttribute('open');
			expect(trigger).toHaveFocus();
		} finally {
			trigger.remove();
		}
	});
});

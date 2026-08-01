import { fireEvent, render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import DirectMessageChooser from './DirectMessageChooser.svelte';

describe('DirectMessageChooser', () => {
	it('names the dialog and chooses a player through a native button', async () => {
		const onchoose = vi.fn();
		render(DirectMessageChooser, {
			props: {
				open: true,
				description: 'Choose a player to start a direct conversation.',
				entries: [
					{
						id: 'participant-2',
						displayLabel: 'Rowan',
						supportingLabel: 'Seat 2',
						avatarText: 'R'
					}
				],
				onchoose,
				close: vi.fn()
			}
		});

		expect(screen.getByRole('dialog', { name: 'New message' })).toHaveAttribute('open');
		const recipient = screen.getByRole('button', { name: /Rowan.*Seat 2/ });
		expect(recipient).toBeInstanceOf(HTMLButtonElement);
		expect(recipient).toHaveTextContent('R');
		await fireEvent.click(recipient);
		expect(onchoose).toHaveBeenCalledWith('participant-2');
	});

	it('explains when no players can be chosen', () => {
		render(DirectMessageChooser, {
			props: {
				open: true,
				description: 'Choose a player to start a direct conversation.',
				entries: [],
				onchoose: vi.fn(),
				close: vi.fn()
			}
		});

		expect(screen.getByText('No players are available for a direct conversation.')).toBeVisible();
	});
});

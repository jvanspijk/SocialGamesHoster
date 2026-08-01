import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it, vi } from 'vitest';
import Toast from './Toast.svelte';

afterEach(cleanup);

describe('Toast', () => {
	it.each(['error', 'success', 'info'] as const)('renders one polite %s notification', (tone) => {
		render(Toast, {
			props: { tone, message: `${tone} message`, ondismiss: vi.fn() }
		});

		const toast = screen.getByRole('status');
		expect(toast).toHaveTextContent(`${tone} message`);
		expect(toast).toHaveAttribute('aria-live', 'polite');
	});

	it('runs its optional action and can be dismissed accessibly', async () => {
		const onaction = vi.fn();
		const ondismiss = vi.fn();
		render(Toast, {
			props: {
				tone: 'error',
				message: 'Requests could not be loaded.',
				actionLabel: 'Retry',
				onaction,
				ondismiss
			}
		});

		await fireEvent.click(screen.getByRole('button', { name: 'Retry' }));
		expect(onaction).toHaveBeenCalledOnce();
		await fireEvent.click(screen.getByRole('button', { name: 'Dismiss notification' }));
		expect(ondismiss).toHaveBeenCalledOnce();
	});
});

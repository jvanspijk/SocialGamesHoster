import { cleanup, fireEvent, render, screen } from '@testing-library/svelte';
import { tick } from 'svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { toasts } from '$lib/state/toasts.svelte';
import ToastViewport from './ToastViewport.svelte';

beforeEach(() => {
	vi.useFakeTimers();
	toasts.clear();
});

afterEach(() => {
	cleanup();
	toasts.clear();
	vi.useRealTimers();
});

describe('ToastViewport', () => {
	it('dismisses a notification from its accessible close control', async () => {
		toasts.info('Settings saved.');
		render(ToastViewport);

		await fireEvent.click(screen.getByRole('button', { name: 'Dismiss notification' }));
		expect(toasts.items).toEqual([]);
	});

	it('removes transient notifications after their timeout', async () => {
		toasts.success('Profile saved.');
		render(ToastViewport);
		await tick();

		await vi.advanceTimersByTimeAsync(4_000);
		expect(toasts.items).toEqual([]);
	});

	it('keeps persistent notifications visible after the normal timeout', async () => {
		toasts.error('Conversations could not be loaded.', { persistent: true });
		render(ToastViewport);
		await tick();

		await vi.advanceTimersByTimeAsync(20_000);
		expect(screen.getByText('Conversations could not be loaded.')).toBeVisible();
		expect(toasts.items).toHaveLength(1);
	});

	it('pauses and resumes the timeout while a notification is hovered', async () => {
		toasts.info('Restore scheduled.');
		render(ToastViewport);
		await tick();
		const toast = screen.getByRole('status');

		await fireEvent.mouseEnter(toast);
		await vi.advanceTimersByTimeAsync(10_000);
		expect(toasts.items).toHaveLength(1);
		await fireEvent.mouseLeave(toast);
		await vi.advanceTimersByTimeAsync(4_000);
		expect(toasts.items).toEqual([]);
	});

	it('pauses and resumes the timeout while a notification contains focus', async () => {
		toasts.info('Restore scheduled.', { actionLabel: 'View', action: vi.fn() });
		render(ToastViewport);
		await tick();
		const toast = screen.getByRole('status');

		await fireEvent.focusIn(toast);
		await vi.advanceTimersByTimeAsync(10_000);
		expect(toasts.items).toHaveLength(1);
		await fireEvent.focusOut(toast);
		await vi.advanceTimersByTimeAsync(4_000);
		expect(toasts.items).toEqual([]);
	});

	it('waits for both hover and focus to end, including a move between toast controls', async () => {
		toasts.info('Restore scheduled.', { actionLabel: 'View', action: vi.fn() });
		render(ToastViewport);
		await tick();
		const toast = screen.getByRole('status');
		const action = screen.getByRole('button', { name: 'View' });
		const dismiss = screen.getByRole('button', { name: 'Dismiss notification' });

		await fireEvent.mouseEnter(toast);
		await fireEvent.focusIn(action);
		await fireEvent.mouseLeave(toast);
		await vi.advanceTimersByTimeAsync(10_000);
		expect(toasts.items).toHaveLength(1);

		await fireEvent.focusOut(action, { relatedTarget: dismiss });
		await fireEvent.focusIn(dismiss, { relatedTarget: action });
		await vi.advanceTimersByTimeAsync(10_000);
		expect(toasts.items).toHaveLength(1);

		await fireEvent.focusOut(dismiss);
		await vi.advanceTimersByTimeAsync(4_000);
		expect(toasts.items).toEqual([]);
	});
});

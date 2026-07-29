import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AppApiError } from '$lib/api/client';
import { toasts } from '$lib/state/toasts.svelte';
import PendingProfileRequests from './PendingProfileRequests.svelte';

const mocks = vi.hoisted(() => ({
	api: vi.fn(),
	subscribe: vi.fn(),
	unsubscribe: vi.fn()
}));

vi.mock('$lib/api/client', async (importOriginal) => {
	const original = await importOriginal<typeof import('$lib/api/client')>();
	return {
		...original,
		api: mocks.api,
		pb: { realtime: { subscribe: mocks.subscribe } }
	};
});

const newRequest = {
	id: 'new-1',
	requestType: 'new' as const,
	requestedName: 'Alice',
	createdAt: '2026-07-29T10:00:00Z',
	expiresAt: '2026-07-29T10:15:00Z'
};

const recoveryRequest = {
	id: 'recover-1',
	requestType: 'recover' as const,
	requestedName: 'Bob',
	createdAt: '2026-07-29T10:01:00Z',
	expiresAt: '2026-07-29T10:16:00Z'
};

beforeEach(() => {
	mocks.api.mockReset();
	mocks.subscribe.mockReset();
	mocks.unsubscribe.mockReset();
	mocks.subscribe.mockResolvedValue(mocks.unsubscribe);
	toasts.clear();
});

afterEach(cleanup);

describe('PendingProfileRequests', () => {
	it('loads requests, reports the count, and cleans up live updates', async () => {
		mocks.api.mockResolvedValueOnce([newRequest, recoveryRequest]);
		const oncountchange = vi.fn();
		const { unmount } = render(PendingProfileRequests, { props: { oncountchange } });

		expect(await screen.findByText('Alice')).toBeInTheDocument();
		expect(screen.getByText('New profile')).toBeInTheDocument();
		expect(screen.getByText('Profile recovery')).toBeInTheDocument();
		expect(oncountchange).toHaveBeenCalledWith(2);
		expect(mocks.subscribe).toHaveBeenCalledWith(
			'profile-requests:game-masters',
			expect.any(Function)
		);
		await waitFor(() => expect(mocks.subscribe).toHaveResolved());

		unmount();
		await waitFor(() => expect(mocks.unsubscribe).toHaveBeenCalledOnce());
	});

	it('approves new requests directly and confirms profile recoveries', async () => {
		mocks.api
			.mockResolvedValueOnce([newRequest, recoveryRequest])
			.mockResolvedValueOnce(undefined)
			.mockResolvedValueOnce(undefined);
		const onapproved = vi.fn();
		render(PendingProfileRequests, { props: { onapproved } });

		await screen.findByText('Alice');
		const approveButtons = screen.getAllByRole('button', { name: 'Approve' });
		await fireEvent.click(approveButtons[0]);

		await waitFor(() =>
			expect(mocks.api).toHaveBeenCalledWith('/admin/profile-requests/new-1/approve', {
				method: 'POST',
				body: {}
			})
		);
		expect(screen.queryByText('Alice')).not.toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Approve' }));
		expect(screen.getByRole('heading', { name: 'Approve profile recovery?' })).toBeInTheDocument();
		expect(
			screen.getByText('The previous device session will stop working immediately.')
		).toBeInTheDocument();
		await fireEvent.click(screen.getByRole('button', { name: 'Approve recovery' }));

		await waitFor(() => expect(onapproved).toHaveBeenCalledTimes(2));
		expect(screen.getByText('No pending requests')).toBeInTheDocument();
	});

	it('rejects with an optional reason and refreshes an already-decided request', async () => {
		mocks.api
			.mockResolvedValueOnce([newRequest])
			.mockRejectedValueOnce(
				new AppApiError({ code: 'request.decided', message: 'Already decided.' }, 409)
			)
			.mockResolvedValueOnce([]);
		render(PendingProfileRequests);

		await screen.findByText('Alice');
		await fireEvent.click(screen.getByRole('button', { name: 'Reject' }));
		await fireEvent.input(screen.getByRole('textbox', { hidden: true }), {
			target: { value: 'Name is unclear' }
		});
		await fireEvent.click(screen.getByRole('button', { name: 'Reject request' }));

		await waitFor(() => expect(screen.getByText('No pending requests')).toBeInTheDocument());
		expect(mocks.api).toHaveBeenCalledWith('/admin/profile-requests/new-1/reject', {
			method: 'POST',
			body: { reason: 'Name is unclear' }
		});
		expect(toasts.items.some((toast) => toast.message.includes('already decided'))).toBe(true);
	});
});

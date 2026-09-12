import { cleanup, render, waitFor } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import ProfileRequestAttention from './ProfileRequestAttention.svelte';

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

const request = {
	id: 'request-1',
	requestType: 'new' as const,
	requestedName: 'Alice',
	createdAt: '2026-09-10T10:00:00Z',
	expiresAt: '2026-09-10T10:10:00Z'
};

beforeEach(() => {
	mocks.api.mockReset();
	mocks.subscribe.mockReset();
	mocks.unsubscribe.mockReset();
	mocks.subscribe.mockResolvedValue(mocks.unsubscribe);
});

afterEach(cleanup);

describe('ProfileRequestAttention', () => {
	it('subscribes before loading and updates the count after a live request event', async () => {
		let resolveSubscription: (() => void) | undefined;
		mocks.subscribe.mockImplementation(
			() =>
				new Promise<() => void>((resolve) => {
					resolveSubscription = () => resolve(mocks.unsubscribe);
				})
		);
		mocks.api.mockResolvedValueOnce([]).mockResolvedValueOnce([request]);
		const oncountchange = vi.fn();
		const { unmount } = render(ProfileRequestAttention, { props: { oncountchange } });

		expect(mocks.subscribe).toHaveBeenCalledWith(
			'profile-requests:game-masters',
			expect.any(Function)
		);
		expect(mocks.api).not.toHaveBeenCalled();

		resolveSubscription?.();
		await waitFor(() => expect(oncountchange).toHaveBeenCalledWith(0));

		const refresh = mocks.subscribe.mock.calls[0][1] as () => Promise<void>;
		await refresh();
		await waitFor(() => expect(oncountchange).toHaveBeenCalledWith(1));

		unmount();
		await waitFor(() => expect(mocks.unsubscribe).toHaveBeenCalledOnce());
	});
});

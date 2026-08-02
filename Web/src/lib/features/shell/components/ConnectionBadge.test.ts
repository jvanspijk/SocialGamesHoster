import { cleanup, render, screen, waitFor } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import { connection } from '$lib/state/connection.svelte';
import ConnectionBadge from './ConnectionBadge.svelte';

afterEach(() => {
	cleanup();
	connection.set('connected');
});

describe('ConnectionBadge', () => {
	it.each([
		['connected', 'connected', 'tone-success'],
		['reconnecting', 'reconnecting', 'tone-warning'],
		['offline', 'offline', 'tone-danger']
	] as const)('maps %s state to its label and tone', async (state, label, tone) => {
		connection.set(state);
		const view = render(ConnectionBadge);

		await waitFor(() => expect(screen.getByText(label)).toBeVisible());
		expect(view.container.querySelector('.status-badge')).toHaveClass(tone);
	});
});

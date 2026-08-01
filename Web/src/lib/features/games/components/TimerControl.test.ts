import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import type { TimerProjection } from '$lib/api/types';
import TimerControl from './TimerControl.svelte';

afterEach(cleanup);

function timer(status: TimerProjection['status']): TimerProjection {
	return {
		status,
		totalMs: 300_000,
		remainingMs: status === 'completed' ? 0 : 120_000,
		revision: 1,
		serverTime: new Date().toISOString()
	};
}

describe('TimerControl', () => {
	it.each([
		['inactive', 'Start timer', ['Pause timer', 'Resume timer']],
		['running', 'Pause timer', ['Start timer', 'Resume timer']],
		['paused', 'Resume timer', ['Start timer', 'Pause timer']],
		['completed', 'Start again', ['Pause timer', 'Resume timer']]
	] as const)('shows only valid actions while %s', (status, expected, absent) => {
		render(TimerControl, {
			props: { gameId: 'game', timer: timer(status), onchange: () => undefined }
		});

		expect(screen.getByRole('button', { name: expected })).toBeInTheDocument();
		for (const label of absent) {
			expect(screen.queryByRole('button', { name: label })).not.toBeInTheDocument();
		}
	});
});

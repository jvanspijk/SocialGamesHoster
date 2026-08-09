import { cleanup, render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Countdown from './Countdown.svelte';
import { countdownAccessibleLabel, formatCountdown, remainingCountdownMs } from './countdown';

describe('countdown formatting', () => {
	it.each([
		[59_999, '00:59'],
		[3_661_000, '61:01'],
		[0, '00:00'],
		[-1, '00:00']
	])('formats %d milliseconds as %s', (remainingMs, expected) => {
		expect(formatCountdown(remainingMs)).toBe(expected);
	});

	it('clamps a completed running countdown at zero', () => {
		expect(
			remainingCountdownMs(
				'running',
				5_000,
				'2026-01-01T00:00:00.000Z',
				Date.UTC(2026, 0, 1, 0, 0, 1)
			)
		).toBe(0);
		expect(countdownAccessibleLabel(0, 'seconds remaining')).toBe('0 seconds remaining');
	});
});

describe('Countdown', () => {
	beforeEach(() => {
		vi.useFakeTimers();
		vi.setSystemTime(new Date('2026-01-01T00:00:00.000Z'));
	});

	afterEach(() => {
		cleanup();
		vi.useRealTimers();
	});

	it('updates a running countdown every 250 milliseconds from its end time', async () => {
		render(Countdown, {
			props: {
				status: 'running',
				statusLabel: 'Timer running',
				remainingMs: 65_000,
				endsAt: '2026-01-01T00:01:05.000Z'
			}
		});

		expect(screen.getByLabelText('65 seconds remaining')).toHaveTextContent('01:05');
		await vi.advanceTimersByTimeAsync(250);
		expect(screen.getByLabelText('65 seconds remaining')).toHaveTextContent('01:04');
	});

	it('keeps a paused countdown at its supplied remaining time', async () => {
		render(Countdown, {
			props: {
				status: 'paused',
				statusLabel: 'Timer paused',
				remainingMs: 65_000,
				endsAt: '2026-01-01T00:00:01.000Z'
			}
		});

		await vi.advanceTimersByTimeAsync(5_000);
		expect(screen.getByLabelText('65 seconds remaining')).toHaveTextContent('01:05');
		expect(screen.getByText('Timer paused')).toBeInTheDocument();
	});

	it('presents a completed countdown at zero with the completed tone', () => {
		const { container } = render(Countdown, {
			props: {
				status: 'completed',
				statusLabel: 'Timer completed',
				remainingMs: 0
			}
		});

		expect(screen.getByLabelText('0 seconds remaining')).toHaveTextContent('00:00');
		expect(container.querySelector('.completed')).toBeInTheDocument();
	});
});

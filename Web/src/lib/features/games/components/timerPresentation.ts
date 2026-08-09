import type { TimerProjection } from '$lib/api/types';
import type { CountdownStatus } from '$lib/components/ui/countdown';

export function countdownStatus(status: TimerProjection['status']): CountdownStatus {
	return (
		{
			inactive: 'idle',
			running: 'running',
			paused: 'paused',
			completed: 'completed'
		} as const
	)[status];
}

export function timerStatusLabel(status: TimerProjection['status']): string {
	return (
		{
			inactive: 'No timer set',
			running: 'Timer running',
			paused: 'Timer paused',
			completed: 'Timer completed'
		} as const
	)[status];
}

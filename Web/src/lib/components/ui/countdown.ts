export type CountdownStatus = 'idle' | 'running' | 'paused' | 'completed';

export function remainingCountdownMs(
	status: CountdownStatus,
	remainingMs: number,
	endsAt: string | undefined,
	now: number
): number {
	if (status !== 'running' || !endsAt) return Math.max(0, remainingMs);

	const endTime = new Date(endsAt).getTime();
	return Number.isNaN(endTime) ? Math.max(0, remainingMs) : Math.max(0, endTime - now);
}

export function formatCountdown(remainingMs: number): string {
	const safeRemainingMs = Math.max(0, remainingMs);
	const minutes = Math.floor(safeRemainingMs / 60_000);
	const seconds = Math.floor((safeRemainingMs % 60_000) / 1_000);
	return `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
}

export function countdownAccessibleLabel(remainingMs: number, label: string): string {
	return `${Math.ceil(Math.max(0, remainingMs) / 1_000)} ${label}`;
}

<script lang="ts">
	import TimeDisplay from './TimeDisplay.svelte';
	import { PauseTimer } from '$lib/client/Timers/PauseTimer';
	import { ResumeTimer } from '$lib/client/Timers/ResumeTimer';

	let { timer } = $props<{
		timer: {
			isRunning: boolean;
			totalSeconds: number;
			remainingSeconds: number;
		};
	}>();

	let isLoading = $state(false);
	let errorMessage = $state<string | null>(null);

	// Bound from TimeDisplay - single source of truth for timer state.
	// PERF: currentTime updates ~60fps, but $derived coalesces to boolean changes only.
	let currentTime = $state(0);
	let isRunning = $state(false);
	const isTimerFinished = $derived(currentTime === 0 && !isRunning);

	async function toggleTimer() {
		if (isTimerFinished || isLoading) return;

		isLoading = true;
		errorMessage = null;

		const result = isRunning
			? await PauseTimer(fetch, undefined)
			: await ResumeTimer(fetch, undefined);

		isLoading = false;

		if (!result.ok) {
			errorMessage = result.error.detail ?? 'Failed to update timer';
			setTimeout(() => {
				errorMessage = null;
			}, 3000);
		}
		// Success: SignalR updates TimeDisplay, which updates our bound state
	}
</script>

<div class="timer-interactive-wrapper" class:non-interactive={isTimerFinished}>
	<TimeDisplay
		initialTotalSeconds={timer.totalSeconds}
		initialRemainingTime={timer.remainingSeconds}
		initialIsRunning={timer.isRunning}
		bind:currentTime
		bind:running={isRunning}
		onFinished={() => {}}
		onclick={!isTimerFinished ? toggleTimer : undefined}
	/>

	{#if !isTimerFinished}
		<div class="pause-overlay" class:loading={isLoading}>
			{#if isLoading}
				<div class="spinner"></div>
			{:else if isRunning}
				<svg viewBox="0 0 24 24" fill="currentColor"
					><path d="M6 19h4V5H6v14zm8-14v14h4V5h-4z" /></svg
				>
			{:else}
				<svg viewBox="0 0 24 24" fill="currentColor"><path d="M8 5v14l11-7z" /></svg>
			{/if}
		</div>
	{/if}

	{#if errorMessage}
		<div class="error-toast">{errorMessage}</div>
	{/if}
</div>

<style>
	.timer-interactive-wrapper {
		position: relative;
		height: 75px;
		display: flex;
		align-items: center;
		max-width: fit-content;
	}

	.non-interactive {
		pointer-events: none;
		cursor: default;
	}

	.pause-overlay {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		display: flex;
		align-items: center;
		justify-content: center;
		background: rgba(0, 0, 0, 0.4);
		border-radius: 50%;
		color: var(--color-surface);
		opacity: 0;
		transition: opacity 0.15s ease-out;
		pointer-events: none;
		z-index: 20;
	}

	.timer-interactive-wrapper:hover .pause-overlay,
	.pause-overlay.loading {
		opacity: 1;
	}

	.pause-overlay svg {
		width: 35px;
		height: 35px;
	}

	.spinner {
		width: 24px;
		height: 24px;
		border: 3px solid rgba(255, 255, 255, 0.3);
		border-top-color: var(--color-surface);
		border-radius: 50%;
		animation: spin 0.8s linear infinite;
	}

	@keyframes spin {
		to {
			transform: rotate(360deg);
		}
	}

	.error-toast {
		position: absolute;
		bottom: -30px;
		left: 50%;
		transform: translateX(-50%);
		background: var(--color-accent-strong);
		color: var(--color-on-accent);
		padding: 4px 10px;
		border-radius: 3px;
		font-size: 0.75rem;
		white-space: nowrap;
		z-index: 30;
		animation: fadeIn 0.15s ease-out;
	}

	@keyframes fadeIn {
		from {
			opacity: 0;
			transform: translateX(-50%) translateY(-5px);
		}
		to {
			opacity: 1;
			transform: translateX(-50%) translateY(0);
		}
	}
</style>

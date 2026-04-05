<script lang="ts">
	import { timersHub } from '$lib/client/Timers/TimersHub.svelte';
	import { onMount } from 'svelte';

	let {
		initialRemainingTime = 0,
		initialTotalSeconds = 0,
		initialIsRunning = false,
		onFinished,
		onclick,
		// Bindable state - allows parent components to react to timer state changes.
		// PERF: displayTime updates at ~60fps when running. Binding it is safe because
		// Svelte only re-renders when $derived values change. Use $derived in the parent
		// to coalesce updates (e.g., `$derived(displayTime < 10)` only triggers on threshold crossing).
		currentTime = $bindable(0),
		running = $bindable(false)
	}: {
		initialRemainingTime: number;
		initialTotalSeconds: number;
		initialIsRunning?: boolean;
		onFinished?: () => void;
		onclick?: () => void;
		currentTime?: number;
		running?: boolean;
	} = $props();

	// --- Server State ---
	let remaining = $state(0);
	let total = $state(0);
	let isRunning = $state(false);
	let lastServerUpdate = $state(Date.now());

	// --- Display State ---
	let displayTime = $state(0);

	// Sync bindable props with internal state
	$effect(() => {
		currentTime = displayTime;
	});
	$effect(() => {
		running = isRunning;
	});

	// Initialize from props on mount. After this, SignalR takes over.
	onMount(() => {
		remaining = initialRemainingTime;
		total = initialTotalSeconds;
		isRunning = initialIsRunning;
		displayTime = initialRemainingTime;
	});

	// Sync display with server state when the timer is not running.
	$effect(() => {
		if (!isRunning) {
			displayTime = remaining;
		}
	});

	// --- SVG Progress Bar Calculation ---
	const R = 45;
	const C = 2 * Math.PI * R;
	const progressOffset = $derived(C - (Math.max(0, displayTime) / (total || 1)) * C);

	const handleSync = (payload: { remainingSeconds: number; totalSeconds: number }) => {
		remaining = payload.remainingSeconds;
		total = payload.totalSeconds;
		isRunning = true;
		lastServerUpdate = Date.now();
		displayTime = remaining;
	};

	timersHub.onEvent('OnTimerStarted', handleSync);
	timersHub.onEvent('OnTimerResumed', handleSync);
	timersHub.onEvent('OnTimerChanged', handleSync);

	timersHub.onEvent('OnTimerPaused', (payload) => {
		remaining = payload.remainingSeconds;
		isRunning = false;
	});

	timersHub.onEvent('OnTimerStopped', () => {
		isRunning = false;
		remaining = 0;
		onFinished?.();
	});

	timersHub.onEvent('OnTimerFinished', () => {
		isRunning = false;
		remaining = 0;
		onFinished?.();
	});

	// --- Interpolation Effect ---
	// This effect runs the client-side animation loop.
	$effect(() => {
		if (!isRunning) return;

		let frame: number;
		const tick = () => {
			const elapsed = (Date.now() - lastServerUpdate) / 1000;
			const interpolated = remaining - elapsed;

			displayTime = Math.max(0, interpolated);

			if (interpolated > 0) {
				frame = requestAnimationFrame(tick);
			}
		};

		frame = requestAnimationFrame(tick);
		return () => cancelAnimationFrame(frame);
	});

	// --- Utility Functions ---
	function formatTime(seconds: number): string {
		const safe = Math.max(0, Math.floor(seconds));
		return `${String(Math.floor(safe / 60)).padStart(2, '0')}:${String(safe % 60).padStart(2, '0')}`;
	}

	const isUrgent = $derived(displayTime < total * 0.25);
</script>

<div class="timer-container" class:urgent={isUrgent} class:inactive={!isRunning}>
	<div class="timer-wrapper">
		<svg class="progress-ring" viewBox="0 0 100 100">
			<circle class="ring-track" cx="50" cy="50" r={R} />
			<circle
				class="ring-progress"
				cx="50"
				cy="50"
				r={R}
				stroke-dasharray={C}
				stroke-dashoffset={progressOffset}
			/>
		</svg>
		<div class="time-text">
			{formatTime(displayTime)}
		</div>

		{#if onclick}
			<button type="button" class="hitbox" onclick={() => onclick?.()} aria-label="Timer action"
			></button>
		{/if}
	</div>
</div>

<style>
	.timer-wrapper {
		position: relative;
		width: 75px;
		height: 75px;
		background-color: var(--color-border);
		border-radius: 50%;
		box-shadow:
			0 0 5px rgba(255, 255, 255, 0.2),
			inset 0 0 15px rgba(0, 0, 0, 0.9);
		border: 2px solid var(--color-border);
		transition:
			transform 0.1s ease-out,
			box-shadow 0.1s ease-out;
		overflow: hidden;
	}

	.timer-wrapper:has(.hitbox:active) {
		transform: scale(0.92);
		box-shadow:
			0 0 2px rgba(255, 255, 255, 0.1),
			inset 0 0 20px rgba(0, 0, 0, 1);
	}

	.hitbox {
		position: absolute;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		z-index: 10;

		appearance: none;
		background: transparent;
		border: none;
		padding: 0;
		margin: 0;
		cursor: pointer;

		border-radius: 50%;

		outline: none;
	}

	.hitbox:hover {
		background: rgba(255, 255, 255, 0.025);
	}

	.hitbox:focus-visible {
		box-shadow: 0 0 0 3px rgba(212, 142, 21, 0.5);
	}

	.inactive {
		opacity: 0.5;
		filter: grayscale(0.6) brightness(0.8);
	}

	.inactive .ring-progress {
		stroke: var(--color-accent-strong);
	}

	.progress-ring {
		position: absolute;
		width: 100%;
		height: 100%;
		transform: rotate(-90deg);
		transition: transform 0.15s ease-out;
	}

	.ring-track {
		fill: none;
		stroke: var(--color-ink-deep);
		stroke-width: 6;
	}

	.ring-progress {
		fill: none;
		stroke: var(--color-accent-strong);
		stroke-width: 10;
		stroke-linecap: round;
		transition: stroke-dashoffset 0.05s linear;
	}

	.time-text {
		position: absolute;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		font-family: var(--font-heading);
		font-size: 1em;
		font-weight: 700;
		color: var(--color-surface);
		user-select: none;
	}

	.urgent .ring-progress {
		stroke: var(--color-warning);
		filter: drop-shadow(0 0 5px var(--color-warning));
	}

	.urgent .time-text {
		color: var(--color-on-accent);
		animation: pulse 1s infinite alternate;
	}

	@keyframes pulse {
		from {
			opacity: 1;
			transform: translate(-50%, -50%) scale(1);
		}
		to {
			opacity: 0.8;
			transform: translate(-50%, -50%) scale(1.05);
		}
	}
</style>

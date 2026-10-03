<script lang="ts">
	import { onMount } from 'svelte';
	import Pause from '@lucide/svelte/icons/pause';
	import Play from '@lucide/svelte/icons/play';
	import TimerReset from '@lucide/svelte/icons/timer-reset';
	import {
		countdownAccessibleLabel,
		formatCountdown,
		remainingCountdownMs,
		type CountdownStatus
	} from './countdown';

	let {
		status,
		statusLabel,
		remainingMs,
		endsAt,
		accessibleLabel = 'seconds remaining',
		variant = 'display',
		compact = false
	}: {
		status: CountdownStatus;
		statusLabel: string;
		remainingMs: number;
		endsAt?: string;
		accessibleLabel?: string;
		variant?: 'display' | 'readout';
		compact?: boolean;
	} = $props();

	let now = $state(Date.now());

	onMount(() => {
		const interval = window.setInterval(() => (now = Date.now()), 250);
		const visibility = () => {
			if (!document.hidden) now = Date.now();
		};
		document.addEventListener('visibilitychange', visibility);
		return () => {
			window.clearInterval(interval);
			document.removeEventListener('visibilitychange', visibility);
		};
	});

	const remaining = $derived(remainingCountdownMs(status, remainingMs, endsAt, now));
	const display = $derived(status === 'idle' ? '--:--' : formatCountdown(remaining));
	const timeLabel = $derived(
		status === 'idle' ? 'Timer inactive' : countdownAccessibleLabel(remaining, accessibleLabel)
	);
</script>

{#if variant === 'display'}
	<div class:compact class:completed={status === 'completed'} class="countdown">
		<span class="icon" aria-hidden="true">
			{#if status === 'running'}<Play size={16} />{:else if status === 'paused'}<Pause
					size={16}
				/>{:else}<TimerReset size={16} />{/if}
		</span>
		<strong aria-label={timeLabel}>{display}</strong>
		<small aria-live="polite"><span>{statusLabel}</span></small>
	</div>
{:else}
	<div class:completed={status === 'completed'} class="countdown-readout">
		<p aria-live="polite">{statusLabel}</p>
		<strong aria-label={timeLabel}>{display}</strong>
	</div>
{/if}

<style>
	.countdown {
		display: inline-grid;
		box-sizing: border-box;
		grid-template-columns: 1rem max-content;
		align-items: center;
		gap: 0.1rem 0.5rem;
		border: 1px solid #9a7e51;
		background: var(--surface-raised);
		padding: 0.55rem 0.8rem;
	}

	.icon {
		grid-row: span 2;
		color: var(--action);
	}

	.countdown strong {
		min-width: 4.5ch;
		font-family: var(--font-display);
		font-size: var(--font-size-lg);
		font-variant-numeric: tabular-nums;
		line-height: 1;
		white-space: nowrap;
	}

	small {
		position: relative;
		color: var(--text-secondary);
		font-size: var(--font-size-sm);
		text-transform: capitalize;
		line-height: 1.5;
		white-space: nowrap;
	}

	/* Keep the original inactive footprint; longer labels can use the padding. */
	small::after {
		content: 'inactive';
		visibility: hidden;
	}

	small span {
		position: absolute;
		inset-inline-start: 50%;
		transform: translateX(-50%);
	}

	.compact {
		border: 0;
		background: transparent;
		padding: 0;
	}

	.completed {
		color: var(--status-danger);
	}

	.countdown-readout p {
		margin: 0;
		color: var(--text-secondary);
		line-height: 1.5;
		white-space: nowrap;
	}

	.countdown-readout strong {
		display: block;
		color: var(--text-primary);
		font-family: var(--font-display);
		font-size: clamp(var(--font-size-2xl), 9vw, 5.5rem);
		font-variant-numeric: tabular-nums;
		letter-spacing: 0.06em;
		line-height: 1;
		white-space: nowrap;
	}

	.countdown-readout.completed strong {
		color: var(--status-danger);
	}
</style>

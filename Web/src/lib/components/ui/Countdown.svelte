<script lang="ts">
	import { onMount } from 'svelte';
	import { Pause, Play, TimerReset } from '@lucide/svelte';
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
	const display = $derived(formatCountdown(remaining));
	const timeLabel = $derived(countdownAccessibleLabel(remaining, accessibleLabel));
</script>

{#if variant === 'display'}
	<div class:compact class:completed={status === 'completed'} class="countdown">
		<span class="icon" aria-hidden="true">
			{#if status === 'running'}<Play size={16} />{:else if status === 'paused'}<Pause
					size={16}
				/>{:else}<TimerReset size={16} />{/if}
		</span>
		<strong aria-label={timeLabel}>{display}</strong>
		<small aria-live="polite">{statusLabel}</small>
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
		grid-template-columns: auto auto;
		align-items: center;
		gap: 0.1rem 0.5rem;
		border: 1px solid #9a7e51;
		background: var(--paper-light);
		padding: 0.55rem 0.8rem;
	}

	.icon {
		grid-row: span 2;
		color: var(--crimson);
	}

	.countdown strong {
		font-family: var(--font-display);
		font-size: 1.25rem;
		font-variant-numeric: tabular-nums;
		letter-spacing: 0.08em;
		line-height: 1;
	}

	small {
		color: var(--ink-soft);
		font-size: 0.7rem;
		text-transform: capitalize;
	}

	.compact {
		border: 0;
		background: transparent;
		padding: 0;
	}

	.completed {
		color: var(--danger);
	}

	.countdown-readout p {
		margin: 0;
		color: var(--ink-soft);
	}

	.countdown-readout strong {
		display: block;
		color: var(--ink);
		font-family: var(--font-display);
		font-size: clamp(2.6rem, 9vw, 5rem);
		font-variant-numeric: tabular-nums;
		letter-spacing: 0.06em;
		line-height: 1;
	}

	.countdown-readout.completed strong {
		color: var(--danger);
	}
</style>

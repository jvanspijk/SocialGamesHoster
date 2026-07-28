<script lang="ts">
	import { onMount } from 'svelte';
	import { CircleStop, Pause, Play, Plus, RotateCcw } from '@lucide/svelte';
	import Button from './Button.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import type { TimerProjection } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';

	let {
		gameId,
		timer,
		onchange
	}: {
		gameId: string;
		timer: TimerProjection;
		onchange: (timer: TimerProjection) => void;
	} = $props();

	let durationMinutes = $state(5);
	let now = $state(Date.now());
	let busy = $state(false);

	onMount(() => {
		const interval = window.setInterval(() => (now = Date.now()), 250);
		return () => window.clearInterval(interval);
	});

	const remaining = $derived.by(() => {
		if (timer.status === 'running' && timer.endsAt) {
			return Math.max(0, new Date(timer.endsAt).getTime() - now);
		}
		return timer.remainingMs;
	});
	const display = $derived(
		`${Math.floor(remaining / 60_000)
			.toString()
			.padStart(2, '0')}:${Math.floor((remaining % 60_000) / 1000)
			.toString()
			.padStart(2, '0')}`
	);
	const statusText = $derived(
		(
			{
				inactive: 'No timer set',
				running: 'Timer running',
				paused: 'Timer paused',
				completed: 'Timer completed'
			} as const
		)[timer.status]
	);

	async function command(
		path: 'start' | 'pause' | 'resume' | 'adjust' | 'stop',
		body: unknown = {}
	) {
		busy = true;
		try {
			const updated = await api<TimerProjection>(`/games/${gameId}/timer/${path}`, {
				method: 'POST',
				...jsonBody(body)
			});
			onchange(updated);
		} catch (caught) {
			toasts.error(caught instanceof Error ? caught.message : 'The timer could not be updated.');
		} finally {
			busy = false;
		}
	}
</script>

<section class:completed={timer.status === 'completed'} aria-label="Game timer">
	<div class="timer-readout">
		<p>{statusText}</p>
		<strong aria-label={`${Math.ceil(remaining / 1000)} seconds remaining`}>{display}</strong>
	</div>

	{#if timer.status === 'inactive'}
		<div class="timer-start">
			<label>
				<span>Duration</span>
				<select bind:value={durationMinutes}>
					<option value={1}>1 minute</option>
					<option value={3}>3 minutes</option>
					<option value={5}>5 minutes</option>
					<option value={10}>10 minutes</option>
					<option value={15}>15 minutes</option>
					<option value={30}>30 minutes</option>
				</select>
			</label>
			<Button
				loading={busy}
				onclick={() => command('start', { durationMs: durationMinutes * 60_000 })}
			>
				<Play size={18} /> Start timer
			</Button>
		</div>
	{:else if timer.status === 'running'}
		<div class="timer-actions">
			<Button loading={busy} onclick={() => command('pause')}
				><Pause size={18} /> Pause timer</Button
			>
			<Button
				variant="secondary"
				disabled={busy}
				onclick={() => command('adjust', { deltaMs: 60_000 })}
			>
				<Plus size={18} /> 1 min
			</Button>
			<Button variant="ghost" disabled={busy} onclick={() => command('stop')}
				><CircleStop size={18} /> Clear</Button
			>
		</div>
	{:else if timer.status === 'paused'}
		<div class="timer-actions">
			<Button loading={busy} onclick={() => command('resume')}
				><Play size={18} /> Resume timer</Button
			>
			<Button
				variant="secondary"
				disabled={busy}
				onclick={() => command('adjust', { deltaMs: 60_000 })}
			>
				<Plus size={18} /> 1 min
			</Button>
			<Button variant="ghost" disabled={busy} onclick={() => command('stop')}
				><CircleStop size={18} /> Clear</Button
			>
		</div>
	{:else}
		<div class="timer-actions">
			<Button
				loading={busy}
				onclick={() =>
					command('start', { durationMs: Math.max(timer.totalMs, durationMinutes * 60_000) })}
			>
				<RotateCcw size={18} /> Start again
			</Button>
			<Button variant="ghost" disabled={busy} onclick={() => command('stop')}
				><CircleStop size={18} /> Clear</Button
			>
		</div>
	{/if}
</section>

<style>
	section {
		display: grid;
		grid-template-columns: minmax(10rem, 1fr) minmax(16rem, auto);
		align-items: center;
		gap: var(--space-4);
	}

	.timer-readout p {
		margin: 0;
		color: var(--ink-soft);
	}

	.timer-readout strong {
		display: block;
		color: var(--ink);
		font-family: var(--font-display);
		font-size: clamp(2.6rem, 9vw, 5rem);
		font-variant-numeric: tabular-nums;
		letter-spacing: 0.06em;
		line-height: 1;
	}

	.completed .timer-readout strong {
		color: var(--danger);
	}

	.timer-start,
	.timer-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		justify-content: flex-end;
		gap: var(--space-2);
	}

	label {
		display: grid;
		gap: var(--space-1);
	}

	label span {
		font-family: var(--font-display);
		font-size: 0.68rem;
		font-weight: 700;
		text-transform: uppercase;
	}

	select {
		min-height: var(--target-size);
		border: var(--border-subtle);
		background: var(--paper-light);
		color: var(--ink);
		padding: var(--space-2);
	}

	@media (max-width: 47.99rem) {
		section {
			grid-template-columns: 1fr;
		}

		.timer-readout {
			text-align: center;
		}

		.timer-start,
		.timer-actions {
			display: grid;
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.timer-start label {
			grid-column: 1 / -1;
		}

		.timer-start :global(button:first-of-type) {
			grid-column: 1 / -1;
		}

		.timer-actions :global(button:first-child) {
			grid-column: 1 / -1;
		}
	}
</style>

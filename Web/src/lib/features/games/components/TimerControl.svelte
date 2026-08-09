<script lang="ts">
	import { CircleStop, Pause, Play, Plus, RotateCcw } from '@lucide/svelte';
	import Button from '$lib/components/Button.svelte';
	import SelectField from '$lib/components/SelectField.svelte';
	import Countdown from '$lib/components/ui/Countdown.svelte';
	import { api, jsonBody } from '$lib/api/client';
	import { errorMessage } from '$lib/api/errors';
	import type { TimerProjection } from '$lib/api/types';
	import { toasts } from '$lib/state/toasts.svelte';
	import { countdownStatus, timerStatusLabel } from './timerPresentation';

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
	let busy = $state(false);
	const durationOptions = [
		{ value: 1, label: '1 minute' },
		{ value: 3, label: '3 minutes' },
		{ value: 5, label: '5 minutes' },
		{ value: 10, label: '10 minutes' },
		{ value: 15, label: '15 minutes' },
		{ value: 30, label: '30 minutes' }
	];

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
			toasts.error(errorMessage(caught, 'The timer could not be updated.'));
		} finally {
			busy = false;
		}
	}
</script>

<section aria-label="Game timer">
	<div class="timer-readout">
		<Countdown
			variant="readout"
			status={countdownStatus(timer.status)}
			statusLabel={timerStatusLabel(timer.status)}
			remainingMs={timer.remainingMs}
			endsAt={timer.endsAt}
		/>
	</div>

	{#if timer.status === 'inactive'}
		<div class="timer-start">
			<SelectField
				label="Duration"
				name="timer-duration"
				bind:value={durationMinutes}
				options={durationOptions}
				disabled={busy}
			/>
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

	.timer-start,
	.timer-actions {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-end;
		justify-content: flex-end;
		gap: var(--space-2);
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

		.timer-start :global(label) {
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
